import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Bus, Train, Ship, MapPin, Calendar, Clock, Printer, ChevronDown } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import API from '../services/api';
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

function TicketCard({ booking, printing, onPrint }) {
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

      <footer className="pf-ticket-foot">
        <span className="pf-demo">Demo ticket · not valid for travel</span>
        <button type="button" className="pf-print" onClick={() => onPrint(booking.id)}>
          <Printer size={14} aria-hidden="true" /> Print
        </button>
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

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    let cancelled = false;
    (async () => {
      try {
        const [p, b] = await Promise.all([API.get('/auth/profile'), API.get('/bookings/my')]);
        if (cancelled) return;
        setProfile(p.data);
        setBookings(b.data || []);
        setLoadedAt(Date.now());
      } catch (err) {
        if (!cancelled) setError(errText(err, 'Could not load your profile.'));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, navigate]);

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
  if (loading) return <p className="pf-state">Loading your profile...</p>;

  return (
    <div className="pf">
      {error && <div className="pf-error" role="alert">{error}</div>}

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
              <TicketCard key={b.id} booking={b} printing={printId === b.id} onPrint={handlePrint} />
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
    </div>
  );
}
