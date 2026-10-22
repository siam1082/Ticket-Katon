import { useEffect, useState, useCallback, useRef } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { CheckCircle2, AlertCircle, Clock, ShieldCheck, X, AlertTriangle, Info } from 'lucide-react';
import API from '../services/api';
import { useAuth } from '../context/AuthContext';
import './SeatSelect.css';

const GATEWAYS = [
  { id: 'BKASH', label: 'bKash', color: '#e2136e' },
  { id: 'NAGAD', label: 'Nagad', color: '#f7941d' },
];

const BUS_SEAT = /^[A-Z]\d+$/;

const errText = (err, fallback) => err.response?.data?.error || fallback;

const formatTimer = (s) =>
  `${String(Math.floor(s / 60)).padStart(2, '0')}:${String(s % 60).padStart(2, '0')}`;

function Seat({ seat, selected, onPick }) {
  const isAvailable = seat.status === 'AVAILABLE';
  const isHeld = seat.status === 'HELD' || seat.status === 'LOCKED';

  let state = 'booked';
  if (selected) state = 'selected';
  else if (isHeld) state = 'held';
  else if (isAvailable) state = 'available';

  const sub = state === 'held' ? 'Held' : state === 'booked' ? 'Booked' : `${seat.fare}৳`;

  return (
    <button
      type="button"
      className={`ss-seat ss-seat--${state}`}
      disabled={!isAvailable}
      aria-pressed={selected}
      aria-label={`Seat ${seat.seat_number}, ${sub}`}
      onClick={() => onPick(seat)}
    >
      <span className="ss-seat-no">{seat.seat_number}</span>
      <span className="ss-seat-sub">{sub}</span>
    </button>
  );
}

function SeatGrid({ seats, selectedNo, onPick }) {
  const render = (s) => (
    <Seat key={s.seat_number} seat={s} selected={selectedNo === s.seat_number} onPick={onPick} />
  );

  // Bus coach: rows A..J, aisle in the middle (2 + 2).
  if (seats.length > 0 && seats.every((s) => BUS_SEAT.test(s.seat_number))) {
    const rows = {};
    seats.forEach((s) => {
      const letter = s.seat_number[0];
      rows[letter] = rows[letter] || [];
      rows[letter].push(s);
    });

    return (
      <div className="ss-bus">
        <div className="ss-bus-front">Front · Driver</div>
        {Object.keys(rows)
          .sort()
          .map((letter) => {
            const row = rows[letter].sort(
              (a, b) => parseInt(a.seat_number.slice(1), 10) - parseInt(b.seat_number.slice(1), 10),
            );
            const half = Math.ceil(row.length / 2);
            return (
              <div key={letter} className="ss-bus-row">
                {row.slice(0, half).map(render)}
                <span className="ss-aisle" aria-hidden="true" />
                {row.slice(half).map(render)}
              </div>
            );
          })}
      </div>
    );
  }

  return <div className="ss-grid">{seats.map(render)}</div>;
}

export default function SeatSelect() {
  const [searchParams] = useSearchParams();
  const tripId = searchParams.get('tripId');
  const navigate = useNavigate();
  const { isAuthenticated, user } = useAuth();

  const [loading, setLoading] = useState(true);
  const [layout, setLayout] = useState(null);
  const [selected, setSelected] = useState(null); // { seat_number, fare }
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  // Checkout state
  const [modal, setModal] = useState(null); // null | 'details' | 'pay'
  const [modalError, setModalError] = useState('');
  const [gateway, setGateway] = useState('BKASH');
  const [passenger, setPassenger] = useState({ name: '', phone: '', gender: 'MALE' });
  const [payment, setPayment] = useState(null); // response of POST /payments/initiate
  const [expiresAt, setExpiresAt] = useState(null);
  const [remaining, setRemaining] = useState(0);
  const [booking, setBooking] = useState(null); // confirmed booking
  const expiredRef = useRef(false);

  const fetchSeats = useCallback(async () => {
    try {
      const res = await API.get(`/trips/${tripId}/seats`);
      setLayout(res.data);
    } catch (err) {
      setError(errText(err, 'Failed to load seats'));
    } finally {
      setLoading(false);
    }
  }, [tripId]);

  useEffect(() => {
    if (!tripId) {
      navigate('/');
      return;
    }
    fetchSeats();
  }, [tripId, navigate, fetchSeats]);

  // Leave the checkout and show a message. The server has already released the seats
  // (or we release them below), so we only refresh the seat map.
  const resetCheckout = useCallback(
    (message, isError = false) => {
      setModal(null);
      setModalError('');
      setPayment(null);
      setSelected(null);
      setExpiresAt(null);
      setNotice(isError ? '' : message || '');
      setError(isError ? message || '' : '');
      fetchSeats();
    },
    [fetchSeats],
  );

  // Close the checkout: cancel the pending payment, or just release the seat hold.
  const closeCheckout = useCallback(
    async (message) => {
      setBusy(true);
      try {
        if (payment) {
          await API.post(`/payments/${payment.payment_id}/simulate`, { result: 'cancel' });
        } else if (selected) {
          await API.post('/seats/release', { trip_id: tripId, seat_numbers: [selected.seat_number] });
        }
      } catch {
        // The hold may already be gone; nothing else to do.
      } finally {
        setBusy(false);
        resetCheckout(message);
      }
    },
    [payment, selected, tripId, resetCheckout],
  );

  // Always call the latest closeCheckout from the timer without restarting it.
  const closeRef = useRef(closeCheckout);
  useEffect(() => {
    closeRef.current = closeCheckout;
  }, [closeCheckout]);

  // Countdown, driven by the expiry time the server gave us. The first tick runs
  // immediately, so the timer never reads a stale "0" when the checkout opens.
  useEffect(() => {
    if (!modal || !expiresAt) return undefined;

    const tick = () => {
      const left = Math.max(0, Math.ceil((expiresAt - Date.now()) / 1000));
      setRemaining(left);
      if (left === 0 && !expiredRef.current) {
        expiredRef.current = true;
        closeRef.current('Your seat reservation expired. Please select your seat again.');
      }
    };

    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [modal, expiresAt]);

  const handleSeatClick = (seat) => {
    if (seat.status !== 'AVAILABLE') return;
    setNotice('');
    setError('');
    setSelected((cur) => (cur?.seat_number === seat.seat_number ? null : { seat_number: seat.seat_number, fare: seat.fare }));
  };

  // Step 1: hold the seat in Redis, then open the checkout.
  const handleProceed = async () => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    if (!selected) return;

    setBusy(true);
    setError('');
    setNotice('');
    try {
      const res = await API.post('/seats/hold', { trip_id: tripId, seat_numbers: [selected.seat_number] });
      const seconds = res.data?.expires_in_seconds || 300;
      setPassenger({
        name: user?.full_name || user?.name || '',
        phone: user?.phone || '',
        gender: 'MALE',
      });
      setModalError('');
      expiredRef.current = false;
      setExpiresAt(Date.now() + seconds * 1000);
      setModal('details');
    } catch (err) {
      setError(errText(err, 'Could not hold this seat. It may have just been taken.'));
      setSelected(null);
      fetchSeats();
    } finally {
      setBusy(false);
    }
  };

  // Step 2: the server creates a pending booking and works out the amount.
  const handleInitiate = async (e) => {
    e.preventDefault();
    setBusy(true);
    setModalError('');
    try {
      const res = await API.post('/payments/initiate', {
        trip_id: tripId,
        seats: [selected.seat_number],
        passengers: [
          {
            name: passenger.name.trim(),
            phone: passenger.phone.trim(),
            gender: passenger.gender,
            seat_number: selected.seat_number,
          },
        ],
        gateway,
      });
      setPayment(res.data);
      // Keep the countdown in step with the server's hold.
      setExpiresAt(Date.now() + res.data.expires_in_seconds * 1000);
      setModal('pay');
    } catch (err) {
      const status = err.response?.status;
      const message = errText(err, 'Could not start the demo payment.');
      if (status === 409) resetCheckout(message, true);
      else setModalError(message);
    } finally {
      setBusy(false);
    }
  };

  // Step 3: choose how the demo payment ends.
  const handleSimulate = async (result) => {
    setBusy(true);
    setModalError('');
    try {
      const res = await API.post(`/payments/${payment.payment_id}/simulate`, { result });
      if (result === 'success') {
        setBooking(res.data.booking);
        setModal(null);
        setPayment(null);
        setExpiresAt(null);
      } else {
        resetCheckout(res.data.message);
      }
    } catch (err) {
      resetCheckout(errText(err, 'The demo payment could not be completed.'), true);
    } finally {
      setBusy(false);
    }
  };

  if (loading) return <p className="ss-loading">Loading seat layout...</p>;

  if (booking) {
    return (
      <div className="ss-page">
        <div className="ss-success">
          <CheckCircle2 size={56} aria-hidden="true" />
          <h2>Demo booking confirmed</h2>
          <p className="ss-muted">This is a simulated ticket. No real money was charged.</p>
          <div className="ss-pnr">
            <span>PNR</span>
            <strong>{booking.pnr_code}</strong>
          </div>
          <ul className="ss-tickets">
            {booking.tickets?.map((t) => (
              <li key={t.id}>
                <span>Seat {t.seat_number} · {t.passenger_name}</span>
                <span className="ss-muted">{t.ticket_number}</span>
              </li>
            ))}
          </ul>
          <p className="ss-total">Total: {booking.total_amount} BDT</p>
          <div className="ss-actions">
            <button type="button" className="btn-primary" onClick={() => navigate('/profile')}>View in My Bookings</button>
            <button type="button" className="btn-secondary" onClick={() => navigate('/')}>Book another</button>
          </div>
        </div>
      </div>
    );
  }

  const seatGroups =
    layout?.seats?.reduce((groups, seat) => {
      const name = seat.deck_or_class || 'General';
      groups[name] = groups[name] || [];
      groups[name].push(seat);
      return groups;
    }, {}) || {};
  const groupCount = Object.keys(seatGroups).length;
  const gw = GATEWAYS.find((g) => g.id === gateway);

  return (
    <div className="ss-page">
      <h2 className="ss-title">Select your seat</h2>
      <p className="ss-muted">
        <strong>{layout?.operator_name}</strong> · Reg: {layout?.registration_no} · {layout?.transport_type}
      </p>

      {error && (
        <div className="ss-alert ss-alert--error" role="alert">
          <AlertCircle size={16} aria-hidden="true" /> <span>{error}</span>
        </div>
      )}
      {notice && (
        <div className="ss-alert ss-alert--info" role="status">
          <Info size={16} aria-hidden="true" /> <span>{notice}</span>
        </div>
      )}

      {Object.entries(seatGroups).map(([title, seats]) => (
        <section key={title} className="ss-group">
          {groupCount > 1 && <h4>{title.replace(/_/g, ' ')}</h4>}
          <SeatGrid seats={seats} selectedNo={selected?.seat_number} onPick={handleSeatClick} />
        </section>
      ))}

      <div className="ss-legend" aria-label="Seat legend">
        <span><i className="ss-dot ss-dot--available" /> Available</span>
        <span><i className="ss-dot ss-dot--selected" /> Selected</span>
        <span><i className="ss-dot ss-dot--held" /> Held (someone is paying)</span>
        <span><i className="ss-dot ss-dot--booked" /> Booked</span>
      </div>

      <div className="ss-bar">
        <div>
          <p className="ss-muted">Selected seat</p>
          <h4>{selected ? `${selected.seat_number} · ${selected.fare} BDT` : 'None'}</h4>
        </div>
        <button type="button" className="btn-primary" onClick={handleProceed} disabled={!selected || busy}>
          {busy && !modal ? 'Holding seat...' : 'Proceed to demo payment'}
        </button>
      </div>

      {modal && (
        <div className="ss-overlay" role="dialog" aria-modal="true" aria-label="Demo payment">
          <div className="ss-modal">
            <div className="ss-modal-head" style={{ background: gw.color }}>
              <div>
                <h3>Demo payment</h3>
                <span>Ticket-Katon sandbox · simulation only</span>
              </div>
              <button type="button" onClick={() => closeCheckout('Checkout cancelled. Your seat was released.')} disabled={busy} aria-label="Cancel and release seat">
                <X size={20} />
              </button>
            </div>

            <div className="ss-modal-warn">
              <AlertTriangle size={16} aria-hidden="true" />
              <span><strong>Simulation only.</strong> Do not enter any real wallet number, PIN or card. None is asked for here.</span>
            </div>

            <div className="ss-modal-timer">
              <span><Clock size={14} aria-hidden="true" /> Reservation expires in</span>
              <strong>{formatTimer(remaining)}</strong>
            </div>

            {modalError && <div className="ss-alert ss-alert--error ss-modal-alert" role="alert">{modalError}</div>}

            {modal === 'details' && (
              <form onSubmit={handleInitiate} className="ss-modal-body">
                <div className="ss-field-row">
                  {GATEWAYS.map((g) => (
                    <button
                      key={g.id}
                      type="button"
                      className={gateway === g.id ? 'ss-gw ss-gw--on' : 'ss-gw'}
                      style={gateway === g.id ? { borderColor: g.color, color: g.color } : undefined}
                      onClick={() => setGateway(g.id)}
                    >
                      {g.label} (demo)
                    </button>
                  ))}
                </div>

                <div className="ss-summary">
                  <span>Seat</span><strong>{selected.seat_number}</strong>
                  <span>Estimated fare</span><strong>{selected.fare} BDT</strong>
                </div>
                <p className="ss-hint">The final amount is calculated by the server.</p>

                <label className="ss-label" htmlFor="pname">Passenger name</label>
                <input id="pname" className="ss-input" value={passenger.name} onChange={(e) => setPassenger({ ...passenger, name: e.target.value })} required minLength={2} />

                <div className="ss-two">
                  <div>
                    <label className="ss-label" htmlFor="pphone">Passenger mobile</label>
                    <input id="pphone" className="ss-input" type="tel" placeholder="017XXXXXXXX" value={passenger.phone} onChange={(e) => setPassenger({ ...passenger, phone: e.target.value })} required />
                  </div>
                  <div>
                    <label className="ss-label" htmlFor="pgender">Gender</label>
                    <select id="pgender" className="ss-input" value={passenger.gender} onChange={(e) => setPassenger({ ...passenger, gender: e.target.value })}>
                      <option value="MALE">Male</option>
                      <option value="FEMALE">Female</option>
                      <option value="OTHER">Other</option>
                    </select>
                  </div>
                </div>

                <button type="submit" className="ss-pay" style={{ background: gw.color }} disabled={busy}>
                  {busy ? 'Starting...' : 'Continue to demo payment'}
                </button>
              </form>
            )}

            {modal === 'pay' && payment && (
              <div className="ss-modal-body">
                <div className="ss-summary">
                  <span>Seat</span><strong>{selected.seat_number}</strong>
                  <span>Gateway</span><strong>{payment.gateway} (demo)</strong>
                  <span>Amount</span><strong>{payment.amount} {payment.currency}</strong>
                </div>
                <p className="ss-hint">Choose how this demo payment should end.</p>

                <button type="button" className="ss-pay" style={{ background: '#16a34a' }} disabled={busy} onClick={() => handleSimulate('success')}>
                  {busy ? 'Processing...' : 'Simulate successful payment'}
                </button>
                <div className="ss-two">
                  <button type="button" className="btn-secondary" disabled={busy} onClick={() => handleSimulate('fail')}>Simulate failure</button>
                  <button type="button" className="btn-secondary" disabled={busy} onClick={() => handleSimulate('cancel')}>Cancel payment</button>
                </div>
              </div>
            )}

            <div className="ss-modal-foot">
              <ShieldCheck size={14} aria-hidden="true" /> Demo sandbox · no real currency is charged
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
