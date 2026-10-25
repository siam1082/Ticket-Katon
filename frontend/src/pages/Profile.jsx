import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Bus, Train, Ship, MapPin, Calendar, Clock, Printer, ChevronDown, X, CheckCircle2 } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import API from '../services/api';
import PageLoader from '../components/PageLoader';
import './Profile.css';

const TRANSPORT = {
  BUS: { label: 'Bus', Icon: Bus, color: '#2563eb' },
  TRAIN: { label: 'Train', Icon: Train, color: '#7c3aed' },
  LAUNCH: { label: 'Launch', Icon: Ship, color: '#0891b2' },
};

const STATUS_LABEL = {
  CONFIRMED: 'Confirmed',
  CANCELLED: 'Cancelled',
  REFUNDED: 'Refunded',
};

const errText = (err, fallback) => err.response?.data?.error || fallback;

const DATE_OPTS = { timeZone: 'Asia/Dhaka', day: '2-digit', month: 'short', year: 'numeric' };
const TIME_OPTS = { timeZone: 'Asia/Dhaka', hour: '2-digit', minute: '2-digit' };

const formatDate = (v) => new Date(v).toLocaleDateString('en-GB', DATE_OPTS);
const formatTime = (v) => new Date(v).toLocaleTimeString('en-GB', TIME_OPTS);

function initials(name) {
  return (name || '?')
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0].toUpperCase())
    .join('');
}

function TicketCard({ booking, printing, onPrint, onCancel }) {
  const t = TRANSPORT[booking.transport_type] || TRANSPORT.BUS;
  const Icon = t.Icon;
  const status = booking.status;

  return (
    <article className={`pf-ticket${printing ? ' printing' : ''}`} style={{ '--tc': t.color }}>
      <header className="pf-ticket-head">
        <span className="pf-type"><Icon size={16} aria-hidden="true" /> {t.label}</span>
        <span className={`pf-status pf-status--${status.toLowerCase()}`}>{STATUS_LABEL[status] || status}</span>
      </header>

      <h4>{booking.operator_name}</h4>
      <p className="pf-pnr">PNR <strong>{booking.pnr_code}</strong></p>

      <div className="pf-route">
        <MapPin size={14} aria-hidden="true" />
        <strong>{booking.origin}</strong> → <strong>{booking.destination}</strong>
      </div>

      <dl className="pf-facts">
        <div><dt>Seats</dt><dd>{booking.seats.join(', ') || '-'}</dd></div>
        <div><dt>Total fare</dt><dd>{booking.total_amount} BDT</dd></div>
        <div><dt>Vehicle</dt><dd>{booking.registration_no}</dd></div>
      </dl>

      <div className="pf-when">
        <span><Calendar size={14} aria-hidden="true" /> {formatDate(booking.departure_time)}</span>
        <span><Clock size={14} aria-hidden="true" /> Departs {formatTime(booking.departure_time)}</span>
        <span>Arrives {formatTime(booking.arrival_time)}</span>
      </div>

      {status === 'CANCELLED' && (
        <p className="pf-refund">
          {booking.refund_amount > 0
            ? `Cancelled. Demo refund: ${booking.refund_amount} BDT`
            : 'Cancelled. No refund applied.'}
        </p>
      )}

      <footer className="pf-ticket-foot">
        <span className="pf-demo">Demo ticket · not valid for travel</span>
        <span className="pf-actions">
          {onCancel && (
            <button type="button" className="pf-cancel" onClick={() => onCancel(booking)}>Cancel ticket</button>
          )}
          <button type="button" className="pf-print" onClick={() => onPrint(booking.id)}>
            <Printer size={14} aria-hidden="true" /> Print
          </button>
        </span>
      </footer>
    </article>
  );
}

export default function Profile() {
  const { isAuthenticated, role } = useAuth();
  const navigate = useNavigate();

  const [profile, setProfile] = useState(null);
  const [bookings, setBookings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showPast, setShowPast] = useState(false);
  const [printId, setPrintId] = useState(null);
  const [loadedAt, setLoadedAt] = useState(0); // time the bookings were fetched

  const load = useCallback(async () => {
    try {
      const [p, b] = await Promise.all([API.get('/auth/profile'), API.get('/bookings/my')]);
      setProfile(p.data);
      setBookings(b.data || []);
      setLoadedAt(Date.now());
      setError('');
    } catch (err) {
      setError(errText(err, 'Could not load your profile.'));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    load();
  }, [isAuthenticated, navigate, load]);

  // ---- cancelling a ticket
  const [cancelTarget, setCancelTarget] = useState(null); // the booking being cancelled
  const [preview, setPreview] = useState(null);
  const [cancelError, setCancelError] = useState('');
  const [cancelling, setCancelling] = useState(false);
  const [notice, setNotice] = useState('');

  const openCancel = async (booking) => {
    setCancelTarget(booking);
    setPreview(null);
    setCancelError('');
    try {
      const res = await API.get(`/bookings/${booking.pnr_code}/cancellation`);
      setPreview(res.data);
    } catch (err) {
      setCancelError(errText(err, 'Could not check the cancellation rules.'));
    }
  };

  const closeCancel = () => {
    if (!cancelling) setCancelTarget(null);
  };

  const confirmCancel = async () => {
    setCancelling(true);
    setCancelError('');
    try {
      const res = await API.post(`/bookings/${cancelTarget.pnr_code}/cancel`);
      setNotice(res.data.message);
      setCancelTarget(null);
      await load();
    } catch (err) {
      setCancelError(errText(err, 'Could not cancel the ticket.'));
    } finally {
      setCancelling(false);
    }
  };

  const { upcoming, past } = useMemo(() => {
    const isUpcoming = (b) => b.status === 'CONFIRMED' && new Date(b.departure_time).getTime() > loadedAt;
    return {
      upcoming: bookings.filter(isUpcoming).sort((a, b) => new Date(a.departure_time) - new Date(b.departure_time)),
      past: bookings.filter((b) => !isUpcoming(b)).sort((a, b) => new Date(b.departure_time) - new Date(a.departure_time)),
    };
  }, [bookings, loadedAt]);

  const handlePrint = (id) => {
    setPrintId(id);
    setTimeout(() => {
      window.print();
      setPrintId(null);
    }, 50);
  };

  if (!isAuthenticated) return null;
  if (loading) return <PageLoader label="Loading your profile…" />;

  return (
    <div className="pf">
      {error && <div className="pf-error" role="alert">{error}</div>}
      {notice && (
        <div className="pf-notice" role="status">
          <CheckCircle2 size={16} aria-hidden="true" /> <span>{notice}</span>
          <button type="button" aria-label="Dismiss" onClick={() => setNotice('')}><X size={14} /></button>
        </div>
      )}

      {profile && (
        <section className="pf-card pf-head">
          <div className="pf-avatar" aria-hidden="true">{initials(profile.full_name)}</div>
          <div className="pf-who">
            <h2>{profile.full_name}</h2>
            <p>{profile.email}</p>
            <p>{profile.phone}</p>
          </div>
          <span className="pf-role">{role === 'admin' ? 'Admin' : 'Passenger'}</span>
        </section>
      )}

      <section className="pf-card">
        <h3>Upcoming trips</h3>
        {upcoming.length === 0 ? (
          <p className="pf-empty">
            No upcoming trips.{' '}
            <button type="button" className="pf-link" onClick={() => navigate('/')}>Book a demo ticket</button>
          </p>
        ) : (
          <div className="pf-grid">
            {upcoming.map((b) => (
              <TicketCard key={b.id} booking={b} printing={printId === b.id} onPrint={handlePrint} onCancel={openCancel} />
            ))}
          </div>
        )}
      </section>

      {past.length > 0 && (
        <section className="pf-card">
          <button type="button" className="pf-toggle" aria-expanded={showPast} onClick={() => setShowPast((v) => !v)}>
            <span>Past trips ({past.length})</span>
            <ChevronDown size={18} className={showPast ? 'flip' : ''} aria-hidden="true" />
          </button>
          {showPast && (
            <div className="pf-grid">
              {past.map((b) => (
                <TicketCard key={b.id} booking={b} printing={printId === b.id} onPrint={handlePrint} />
              ))}
            </div>
          )}
        </section>
      )}

      {cancelTarget && (
        <div className="pf-overlay" role="dialog" aria-modal="true" aria-label="Cancel ticket">
          <div className="pf-modal">
            <header>
              <h3>Cancel this ticket?</h3>
              <button type="button" onClick={closeCancel} aria-label="Close" disabled={cancelling}><X size={18} /></button>
            </header>

            <p className="pf-modal-route">
              {cancelTarget.origin} → {cancelTarget.destination} · PNR {cancelTarget.pnr_code}
            </p>

            {!preview && !cancelError && <p className="pf-muted">Checking the cancellation rules...</p>}
            {cancelError && <div className="pf-error" role="alert">{cancelError}</div>}

            {preview && (
              <>
                <div className="pf-refund-box">
                  <div><span>Ticket price</span><strong>{preview.total_amount} BDT</strong></div>
                  <div><span>Refund ({preview.refund_percent}%)</span><strong>{preview.refund_amount} BDT</strong></div>
                </div>

                {!preview.can_cancel && <div className="pf-error" role="alert">{preview.reason}</div>}

                <ul className="pf-policy">
                  {preview.policy.map((rule) => <li key={rule}>{rule}</li>)}
                </ul>
                <p className="pf-muted">Demo only. No real money is refunded.</p>
              </>
            )}

            <div className="pf-modal-actions">
              <button type="button" className="btn-secondary" onClick={closeCancel} disabled={cancelling}>Keep ticket</button>
              <button type="button" className="pf-danger" onClick={confirmCancel} disabled={!preview?.can_cancel || cancelling}>
                {cancelling ? 'Cancelling...' : 'Cancel ticket'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
