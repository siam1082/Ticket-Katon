import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ShieldCheck, PlusCircle, Ticket, Layers, CheckCircle2, AlertCircle } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import API from '../services/api';
import './AdminDashboard.css';

const EMPTY_FORM = { vehicleId: '', routeId: '', departure: '', arrival: '', fare: '' };

const errText = (err, fallback) => err.response?.data?.error || fallback;

function toLocalInput(date) {
  const d = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return d.toISOString().slice(0, 16);
}

function formatDhaka(value) {
  return new Date(value).toLocaleString('en-GB', {
    timeZone: 'Asia/Dhaka',
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export default function AdminDashboard() {
  const { user, isAuthenticated, role } = useAuth();
  const navigate = useNavigate();
  const allowed = isAuthenticated && role === 'admin';

  const [activeTab, setActiveTab] = useState('trips'); // 'trips' | 'issue'
  const [vehicles, setVehicles] = useState([]);
  const [routes, setRoutes] = useState([]);
  const [trips, setTrips] = useState([]);
  const [form, setForm] = useState(EMPTY_FORM);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  const load = useCallback(async () => {
    try {
      const [v, r, t] = await Promise.all([
        API.get('/admin/vehicles'),
        API.get('/admin/routes'),
        API.get('/admin/trips'),
      ]);
      setVehicles(v.data || []);
      setRoutes(r.data || []);
      setTrips(t.data || []);
      setError('');
    } catch (err) {
      setError(errText(err, 'Could not load admin data.'));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!allowed) {
      navigate('/login');
      return;
    }
    load();
  }, [allowed, navigate, load]);

  if (!allowed) return null;

  const setField = (name) => (e) => setForm({ ...form, [name]: e.target.value });

  const handleCreate = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const res = await API.post('/admin/trips', {
        vehicle_id: form.vehicleId,
        route_id: form.routeId,
        departure_time: form.departure,
        arrival_time: form.arrival,
        base_fare: Number(form.fare),
      });
      setMessage(`Trip created with ${res.data.seats_created} seats.`);
      setForm(EMPTY_FORM);
      await load();
    } catch (err) {
      setError(errText(err, 'Could not create the trip.'));
    } finally {
      setSaving(false);
    }
  };

  const roleLabel = user?.role === 'SUPER_ADMIN' ? 'SUPER ADMIN' : 'OPERATOR ADMIN';
  const minTime = toLocalInput(new Date());

  return (
    <div className="ad">
      <header className="ad-header">
        <div className="ad-brand">
          <span className="ad-logo">TK</span>
          <div>
            <strong>TicketKaton</strong>
            <span>Demo management portal</span>
          </div>
        </div>
        <div className="ad-user">
          <span className="ad-role"><ShieldCheck size={14} aria-hidden="true" /> {roleLabel}</span>
          <span>{user?.full_name || user?.name}</span>
        </div>
      </header>

      <div className="ad-body">
        <nav className="ad-side card" aria-label="Admin sections">
          <button type="button" className={activeTab === 'trips' ? 'on' : ''} onClick={() => setActiveTab('trips')}>
            <PlusCircle size={16} aria-hidden="true" /> Trips
          </button>
          <button type="button" className={activeTab === 'issue' ? 'on' : ''} onClick={() => setActiveTab('issue')}>
            <Ticket size={16} aria-hidden="true" /> Counter ticket
          </button>
          <button type="button" onClick={() => navigate('/profile')}>
            <Layers size={16} aria-hidden="true" /> Profile
          </button>
        </nav>

        <main className="ad-main">
          {error && <div className="ad-alert ad-alert--error" role="alert"><AlertCircle size={16} aria-hidden="true" /> {error}</div>}
          {message && <div className="ad-alert ad-alert--ok" role="status"><CheckCircle2 size={16} aria-hidden="true" /> {message}</div>}

          {activeTab === 'trips' ? (
            <>
              <section className="card">
                <h3 className="card-title">Schedule a new trip</h3>
                <p className="muted">The seats are created automatically from the vehicle&apos;s layout.</p>

                <form onSubmit={handleCreate} className="form-grid ad-form">
                  <div className="form-field">
                    <label htmlFor="vehicle">Vehicle</label>
                    <select id="vehicle" value={form.vehicleId} onChange={setField('vehicleId')} required>
                      <option value="">Select a vehicle</option>
                      {vehicles.map((v) => (
                        <option key={v.id} value={v.id}>
                          {v.operator_name} · {v.registration_no} ({v.type})
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="form-field">
                    <label htmlFor="route">Route</label>
                    <select id="route" value={form.routeId} onChange={setField('routeId')} required>
                      <option value="">Select a route</option>
                      {routes.map((r) => (
                        <option key={r.id} value={r.id}>
                          {r.origin_city} → {r.destination_city}
                        </option>
                      ))}
                    </select>
                  </div>

                  <div className="form-field">
                    <label htmlFor="dep">Departure (Bangladesh time)</label>
                    <input id="dep" type="datetime-local" min={minTime} value={form.departure} onChange={setField('departure')} required />
                  </div>

                  <div className="form-field">
                    <label htmlFor="arr">Arrival (Bangladesh time)</label>
                    <input id="arr" type="datetime-local" min={form.departure || minTime} value={form.arrival} onChange={setField('arrival')} required />
                  </div>

                  <div className="form-field">
                    <label htmlFor="fare">Base fare (BDT)</label>
                    <input id="fare" type="number" min="1" max="100000" step="1" value={form.fare} onChange={setField('fare')} placeholder="e.g. 1200" required />
                  </div>

                  <div className="ad-actions">
                    <button type="submit" className="btn-primary" disabled={saving || loading}>
                      {saving ? 'Creating...' : 'Create trip'}
                    </button>
                    <button type="button" className="btn-secondary" onClick={() => setForm(EMPTY_FORM)}>Reset</button>
                  </div>
                </form>
              </section>

              <section className="card ad-list">
                <h3 className="card-title">Upcoming trips</h3>
                {loading ? (
                  <p className="muted">Loading...</p>
                ) : trips.length === 0 ? (
                  <p className="muted">No upcoming trips.</p>
                ) : (
                  <div className="ad-table-wrap">
                    <table className="ad-table">
                      <thead>
                        <tr>
                          <th>Route</th>
                          <th>Vehicle</th>
                          <th>Departure</th>
                          <th>Fare</th>
                          <th>Seats booked</th>
                          <th>Status</th>
                        </tr>
                      </thead>
                      <tbody>
                        {trips.map((t) => (
                          <tr key={t.id}>
                            <td>{t.origin} → {t.destination}</td>
                            <td>{t.operator_name}<br /><span className="muted">{t.registration_no}</span></td>
                            <td>{formatDhaka(t.departure_time)}</td>
                            <td>{t.base_fare} BDT</td>
                            <td>{t.booked_seats} / {t.total_seats}</td>
                            <td>{t.status}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </section>
            </>
          ) : (
            <section className="card">
              <h3 className="card-title">Counter ticket</h3>
              <p className="muted">
                Counter ticket issuing is not part of this demo build yet. Customers book through the
                normal search, seat and demo payment flow.
              </p>
            </section>
          )}
        </main>
      </div>
    </div>
  );
}
