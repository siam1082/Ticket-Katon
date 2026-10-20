import { useEffect, useState, useCallback } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import API from '../services/api';
import { useAuth } from '../context/AuthContext';
import { CheckCircle2, AlertCircle, Clock, ShieldCheck, X, AlertTriangle } from 'lucide-react';

export default function SeatSelect() {
  const [searchParams] = useSearchParams();
  const tripId = searchParams.get('tripId');
  const navigate = useNavigate();
  const { isAuthenticated, user } = useAuth();

  const [loading, setLoading] = useState(true);
  const [layout, setLayout] = useState(null);
  const [selectedSeat, setSelectedSeat] = useState(null);
  const [selectedSeatFare, setSelectedSeatFare] = useState(0);
  const [isHolding, setIsHolding] = useState(false);
  const [bookingSuccess, setBookingSuccess] = useState(null);
  const [error, setError] = useState('');

  // Payment Modal States
  const [showPaymentModal, setShowPaymentModal] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState('bkash');
  const [walletNumber, setWalletNumber] = useState('');
  const [pinNumber, setPinNumber] = useState('');
  const [holdTimer, setHoldTimer] = useState(300); // 5 minutes TTL match

  const fetchSeats = useCallback(async () => {
    try {
      const res = await API.get(`/trips/${tripId}/seats`);
      setLayout(res.data);
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to load seats');
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

  // Modal Cancel korle instant Redis lock release kora
  const handleClosePaymentModal = async () => {
    try {
      if (selectedSeat) {
        await API.post('/seats/release', {
          trip_id: tripId,
          seat_numbers: [selectedSeat]
        });
      }
    } catch (err) {
      console.error('Failed to release seat lock', err);
    } finally {
      setShowPaymentModal(false);
      setSelectedSeat(null);
      setError('');
      fetchSeats();
    }
  };

  // Hold Timer logic (Synced with 5-minute Redis LockTTL)
  useEffect(() => {
    let timer;
    if (showPaymentModal && holdTimer > 0) {
      timer = setInterval(() => {
        setHoldTimer((prev) => prev - 1);
      }, 1000);
    } else if (holdTimer === 0 && showPaymentModal) {
      alert('Seat reservation expired! Please select your seat again.');
      handleClosePaymentModal();
    }
    return () => clearInterval(timer);
  }, [showPaymentModal, holdTimer]);

  const handleSeatClick = (seat) => {
    if (seat.status !== 'AVAILABLE') return;
    if (selectedSeat === seat.seat_number) {
      setSelectedSeat(null);
      setSelectedSeatFare(0);
    } else {
      setSelectedSeat(seat.seat_number);
      setSelectedSeatFare(seat.fare || 800);
    }
    setError('');
  };

  const handleProceedToPayment = async () => {
    if (!isAuthenticated) {
      alert('Please login to reserve seats and complete payment!');
      navigate('/login');
      return;
    }

    if (!selectedSeat) {
      setError('Please choose a seat first');
      return;
    }

    setIsHolding(true);
    setError('');

    try {
      const holdRes = await API.post('/seats/hold', {
        trip_id: tripId,
        seat_numbers: [selectedSeat]
      });

      const expirySeconds = holdRes.data?.expires_in_seconds || 300;
      setHoldTimer(expirySeconds);
      setShowPaymentModal(true);
    } catch (err) {
      setError(err.response?.data?.error || 'Could not lock seat. It might have just been taken.');
      fetchSeats();
    } finally {
      setIsHolding(false);
    }
  };

  const handleConfirmPayment = async (e) => {
    e.preventDefault();
    if (!walletNumber || !pinNumber) {
      alert('Please enter valid mobile wallet and PIN');
      return;
    }

    setIsHolding(true);
    setError('');

    const demoTxnId = `DEMO-${paymentMethod.toUpperCase()}-${Math.floor(100000 + Math.random() * 900000)}`;

    try {
      const confirmRes = await API.post('/bookings/confirm', {
        trip_id: tripId,
        seats: [selectedSeat],
        passengers: [{
          name: user?.name || "Passenger",
          phone: walletNumber || user?.phone || "01700000000",
          gender: "MALE",
          seat_number: selectedSeat
        }],
        payment: {
          gateway: paymentMethod.toUpperCase(),
          transaction_id: demoTxnId
        }
      });

      setShowPaymentModal(false);
      setBookingSuccess(confirmRes.data);
    } catch (err) {
      setError(err.response?.data?.error || 'Payment confirmation failed or lock expired');
      setShowPaymentModal(false);
      fetchSeats();
    } finally {
      setIsHolding(false);
    }
  };

  const formatTimer = (seconds) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  if (loading) return <p style={{ textAlign: 'center', padding: '60px' }}>Loading seat layout...</p>;

  // Group seats by Coach / Deck if available
  const seatGroups = layout?.seats?.reduce((groups, seat) => {
    const groupName = seat.deck_or_class || 'General';
    if (!groups[groupName]) groups[groupName] = [];
    groups[groupName].push(seat);
    return groups;
  }, {}) || {};

  return (
    <div style={{ maxWidth: '860px', margin: '20px auto', padding: '24px', background: '#fff', borderRadius: '12px', boxShadow: '0 4px 12px rgba(0,0,0,0.05)' }}>
      {bookingSuccess ? (
        <div style={{ textAlign: 'center', padding: '40px 20px' }}>
          <CheckCircle2 size={64} color="#16a34a" style={{ margin: '0 auto 16px' }} />
          <h2 style={{ fontSize: '24px', fontWeight: '800', color: '#16a34a' }}>Payment & Booking Confirmed!</h2>
          <p style={{ marginTop: '8px', color: '#4b5563' }}>Transaction ID: <strong>{bookingSuccess.payment?.transaction_id || 'DEMO-TXN-SUCCESS'}</strong></p>
          <div style={{ margin: '20px auto', display: 'inline-block', background: '#f0fdf4', border: '1px dashed #16a34a', padding: '12px 24px', borderRadius: '8px' }}>
            <span style={{ fontSize: '14px', color: '#166534' }}>PNR Code: </span>
            <strong style={{ fontSize: '20px', letterSpacing: '1px', color: '#15803d' }}>{bookingSuccess.pnr_code || bookingSuccess.pnr}</strong>
          </div>
          <div>
            <button onClick={() => navigate('/profile')} style={{ background: '#16a34a', color: '#fff', border: 'none', padding: '10px 20px', borderRadius: '6px', fontWeight: '600', cursor: 'pointer', marginRight: '10px' }}>
              View in My Bookings
            </button>
            <button onClick={() => navigate('/')} style={{ background: '#f3f4f6', color: '#374151', border: 'none', padding: '10px 20px', borderRadius: '6px', fontWeight: '600', cursor: 'pointer' }}>
              Book Another
            </button>
          </div>
        </div>
      ) : (
        <>
          <h2 style={{ fontSize: '22px', fontWeight: '700', marginBottom: '6px' }}>Select Your Seat</h2>
          <p style={{ color: '#6b7280', marginBottom: '20px' }}>
            <strong>{layout?.operator_name}</strong> • Reg: {layout?.registration_no} • {layout?.transport_type}
          </p>

          {error && (
            <div style={{ background: '#fee2e2', color: '#b91c1c', padding: '12px', borderRadius: '8px', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
              <AlertCircle size={16} />
              <span>{error}</span>
            </div>
          )}

          {/* Seat Layout View (Supports Bus, Train Coach, and Launch Cabins) */}
          {Object.entries(seatGroups).map(([groupTitle, seatsInGroup]) => (
            <div key={groupTitle} style={{ marginBottom: '28px' }}>
              {Object.keys(seatGroups).length > 1 && (
                <h4 style={{ fontSize: '15px', fontWeight: '700', color: '#374151', borderBottom: '1px solid #e5e7eb', paddingBottom: '6px', marginBottom: '16px' }}>
                  {groupTitle.replace('_', ' ')}
                </h4>
              )}

              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(68px, 1fr))', gap: '12px', padding: '16px', background: '#f9fafb', borderRadius: '12px' }}>
                {seatsInGroup.map((s) => {
                  const isAvailable = s.status === 'AVAILABLE';
                  const isHeld = s.status === 'HELD' || s.status === 'LOCKED';
                  const isSelected = selectedSeat === s.seat_number;

                  let bg = '#ffffff';
                  let color = '#1f2937';
                  let border = '1px solid #d1d5db';

                  if (isSelected) {
                    bg = '#22c55e';
                    color = '#ffffff';
                    border = '2px solid #15803d';
                  } else if (isHeld) {
                    bg = '#fef08a';
                    color = '#854d0e';
                    border = '1px solid #eab308';
                  } else if (!isAvailable) {
                    bg = '#9ca3af';
                    color = '#ffffff';
                  }

                  return (
                    <button
                      key={s.seat_number}
                      disabled={!isAvailable}
                      onClick={() => handleSeatClick(s)}
                      style={{
                        height: '54px',
                        borderRadius: '8px',
                        border: border,
                        background: bg,
                        color: color,
                        fontWeight: '700',
                        cursor: isAvailable ? 'pointer' : 'not-allowed',
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        justifyContent: 'center',
                        boxShadow: isSelected ? '0 0 10px rgba(34, 197, 94, 0.4)' : 'none',
                        transition: 'all 0.15s'
                      }}
                    >
                      <span style={{ fontSize: '13px' }}>{s.seat_number}</span>
                      <span style={{ fontSize: '10px' }}>
                        {isHeld ? 'Held' : !isAvailable ? 'Booked' : `${s.fare || 800}৳`}
                      </span>
                    </button>
                  );
                })}
              </div>
            </div>
          ))}

          {/* Legend Identifiers */}
          <div style={{ display: 'flex', justifyContent: 'center', gap: '20px', margin: '24px 0', fontSize: '13px', color: '#4b5563', flexWrap: 'wrap' }}>
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <span style={{ width: '12px', height: '12px', background: '#ffffff', border: '1px solid #d1d5db', marginRight: '6px', borderRadius: '2px' }}></span> Available
            </div>
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <span style={{ width: '12px', height: '12px', background: '#22c55e', marginRight: '6px', borderRadius: '2px' }}></span> Selected
            </div>
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <span style={{ width: '12px', height: '12px', background: '#fef08a', border: '1px solid #eab308', marginRight: '6px', borderRadius: '2px' }}></span> In Payment / Held
            </div>
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <span style={{ width: '12px', height: '12px', background: '#9ca3af', marginRight: '6px', borderRadius: '2px' }}></span> Booked
            </div>
          </div>

          <div style={{ borderTop: '1px solid #e5e7eb', paddingTop: '20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <p style={{ margin: 0, color: '#6b7280', fontSize: '14px' }}>Selected Seat:</p>
              <h4 style={{ margin: 0, fontSize: '18px', fontWeight: '700' }}>{selectedSeat ? `${selectedSeat} (${selectedSeatFare} BDT)` : 'None'}</h4>
            </div>
            <button
              onClick={handleProceedToPayment}
              disabled={!selectedSeat || isHolding}
              style={{
                background: selectedSeat && !isHolding ? '#16a34a' : '#9ca3af',
                color: '#fff',
                border: 'none',
                padding: '12px 28px',
                borderRadius: '8px',
                fontWeight: '700',
                cursor: selectedSeat && !isHolding ? 'pointer' : 'not-allowed'
              }}
            >
              {isHolding ? 'Holding Seat...' : 'Proceed to Payment'}
            </button>
          </div>
        </>
      )}

      {/* Demo Payment Sandbox Modal */}
      {showPaymentModal && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', justifyContent: 'center', alignItems: 'center', zIndex: 1000 }}>
          <div style={{ background: '#fff', width: '100%', maxWidth: '430px', borderRadius: '16px', overflow: 'hidden', boxShadow: '0 20px 25px -5px rgba(0,0,0,0.1)' }}>
            
            <div style={{ background: paymentMethod === 'bkash' ? '#e2136e' : '#f7941d', padding: '16px 20px', color: '#fff', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h3 style={{ margin: 0, fontSize: '18px', fontWeight: 'bold' }}>Demo Payment Gateway</h3>
                <span style={{ fontSize: '12px', opacity: 0.9 }}>Ticket-Katon Secure Sandbox</span>
              </div>
              <button onClick={handleClosePaymentModal} style={{ background: 'none', border: 'none', color: '#fff', cursor: 'pointer' }} title="Cancel and Release Seat">
                <X size={20} />
              </button>
            </div>

            {/* Strict Sandbox Disclaimer */}
            <div style={{ background: '#fffbeb', borderBottom: '1px solid #fef3c7', padding: '10px 16px', display: 'flex', alignItems: 'center', gap: '8px', fontSize: '12px', color: '#b45309' }}>
              <AlertTriangle size={16} style={{ flexShrink: 0 }} />
              <span><strong>Simulation only:</strong> Do NOT enter your real wallet number or actual PIN.</span>
            </div>

            {/* 5-Minute Reverse Countdown Timer */}
            <div style={{ background: '#fef3c7', padding: '8px 16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '13px', color: '#92400e' }}>
              <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}><Clock size={14} /> Reservation expires in:</span>
              <strong>{formatTimer(holdTimer)}</strong>
            </div>

            {/* Gateway Selector */}
            <div style={{ display: 'flex', padding: '16px 20px 0', gap: '10px' }}>
              <button
                type="button"
                onClick={() => setPaymentMethod('bkash')}
                style={{ flex: 1, padding: '10px', borderRadius: '8px', border: paymentMethod === 'bkash' ? '2px solid #e2136e' : '1px solid #e5e7eb', background: paymentMethod === 'bkash' ? '#fdf2f8' : '#fff', fontWeight: 'bold', color: paymentMethod === 'bkash' ? '#e2136e' : '#374151', cursor: 'pointer' }}
              >
                bKash
              </button>
              <button
                type="button"
                onClick={() => setPaymentMethod('nagad')}
                style={{ flex: 1, padding: '10px', borderRadius: '8px', border: paymentMethod === 'nagad' ? '2px solid #f7941d' : '1px solid #e5e7eb', background: paymentMethod === 'nagad' ? '#fffbeb' : '#fff', fontWeight: 'bold', color: paymentMethod === 'nagad' ? '#f7941d' : '#374151', cursor: 'pointer' }}
              >
                Nagad
              </button>
            </div>

            <form onSubmit={handleConfirmPayment} style={{ padding: '20px' }}>
              <div style={{ background: '#f9fafb', padding: '12px', borderRadius: '8px', marginBottom: '16px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '14px', marginBottom: '6px' }}>
                  <span style={{ color: '#6b7280' }}>Selected Seat:</span>
                  <strong>{selectedSeat}</strong>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '15px' }}>
                  <span style={{ color: '#6b7280' }}>Total Amount:</span>
                  <strong style={{ color: '#111827' }}>{selectedSeatFare} BDT</strong>
                </div>
              </div>

              <div style={{ marginBottom: '14px' }}>
                <label style={{ display: 'block', fontSize: '13px', fontWeight: '600', marginBottom: '6px', color: '#374151' }}>Demo {paymentMethod === 'bkash' ? 'bKash' : 'Nagad'} Number</label>
                <input
                  type="text"
                  placeholder="017XXXXXXXX"
                  value={walletNumber}
                  onChange={(e) => setWalletNumber(e.target.value)}
                  style={{ width: '100%', boxSizing: 'border-box', padding: '10px 12px', borderRadius: '6px', border: '1px solid #d1d5db', fontSize: '14px' }}
                  required
                />
              </div>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ display: 'block', fontSize: '13px', fontWeight: '600', marginBottom: '6px', color: '#374151' }}>Enter PIN (Simulation: any 4 digits)</label>
                <input
                  type="password"
                  placeholder="••••"
                  maxLength={5}
                  value={pinNumber}
                  onChange={(e) => setPinNumber(e.target.value)}
                  style={{ width: '100%', boxSizing: 'border-box', padding: '10px 12px', borderRadius: '6px', border: '1px solid #d1d5db', fontSize: '14px' }}
                  required
                />
              </div>

              <button
                type="submit"
                disabled={isHolding}
                style={{
                  width: '100%',
                  padding: '12px',
                  borderRadius: '8px',
                  border: 'none',
                  background: paymentMethod === 'bkash' ? '#e2136e' : '#f7941d',
                  color: '#fff',
                  fontWeight: 'bold',
                  fontSize: '15px',
                  cursor: isHolding ? 'not-allowed' : 'pointer'
                }}
              >
                {isHolding ? 'Processing Payment...' : `Confirm Pay ${selectedSeatFare} BDT`}
              </button>

              <div style={{ textAlign: 'center', marginTop: '12px', fontSize: '12px', color: '#6b7280', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '4px' }}>
                <ShieldCheck size={14} color="#16a34a" /> Demo Sandbox • No real currency will be charged
              </div>
            </form>

          </div>
        </div>
      )}
    </div>
  );
}