import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ShieldCheck, PlusCircle, Ticket, Layers, CheckCircle2, AlertCircle, MapPin, Bus, Trash2, Building2 } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import API from '../services/api';
import './AdminDashboard.css';

const EMPTY_FORM = { vehicleId: '', routeId: '', departure: '', arrival: '', fare: '' };
const EMPTY_ROUTE = { origin_city: '', destination_city: '' };
const EMPTY_OPERATOR = { name: '', type: 'BUS', contact_phone: '' };
const EMPTY_ADMIN = { operatorId: '', email: '', fullName: '', phone: '', password: '' };
const EMPTY_COACH = { coach_name: '', class: 'SNIGDHA', layout: '2x2', total_seats: 60 };
const EMPTY_VEHICLE = { operatorId: '', registrationNo: '', busLayout: '2x2', coaches: [{ ...EMPTY_COACH, coach_name: 'KA' }] };

function layoutSummary(v) {
  const cfg = v.seat_layout_config || {};
  if (v.type === 'BUS') return `${cfg.type || '2x2'} seats`;
  if (v.type === 'LAUNCH') return 'Deck + cabins';
  if (v.type === 'TRAIN' && Array.isArray(cfg.coaches)) {
    return cfg.coaches.map((c) => `${c.coach_name} (${c.total_seats})`).join(', ');
  }
  return '-';
}

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
  const isSuper = user?.role === 'SUPER_ADMIN';

  const [activeTab, setActiveTab] = useState('trips'); // 'trips' | 'routes' | 'vehicles' | 'operators' | 'issue'
  const [vehicles, setVehicles] = useState([]);
  const [routes, setRoutes] = useState([]);
  const [trips, setTrips] = useState([]);
  const [operators, setOperators] = useState([]);
  const [routeForm, setRouteForm] = useState(EMPTY_ROUTE);
  const [vehicleForm, setVehicleForm] = useState(EMPTY_VEHICLE);
  const [operatorAdmins, setOperatorAdmins] = useState([]);
  const [operatorForm, setOperatorForm] = useState(EMPTY_OPERATOR);
  const [adminForm, setAdminForm] = useState(EMPTY_ADMIN);
  const [form, setForm] = useState(EMPTY_FORM);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  const load = useCallback(async () => {
    try {
      const [v, r, t, o] = await Promise.all([
        API.get('/admin/vehicles'),
        API.get('/admin/routes'),
        API.get('/admin/trips'),
        API.get('/admin/operators'),
      ]);
      setOperators(o.data || []);
      if (isSuper) {
        // Separate try: if this one fails, the other lists should still load.
        try {
          const a = await API.get('/admin/operator-admins');
          setOperatorAdmins(a.data || []);
        } catch {
          setOperatorAdmins([]);
        }
      }
      setVehicles(v.data || []);
      setRoutes(r.data || []);
      setTrips(t.data || []);
      setError('');
    } catch (err) {
      setError(errText(err, 'Could not load admin data.'));
    } finally {
      setLoading(false);
    }
  }, [isSuper]);

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

  const switchTab = (tab) => {
    setActiveTab(tab);
    setError('');
    setMessage('');
  };

  const handleCreateRoute = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const res = await API.post('/admin/routes', routeForm);
      setMessage(`Route added: ${res.data.origin_city} → ${res.data.destination_city}. You can now schedule trips on it.`);
      setRouteForm(EMPTY_ROUTE);
      await load();
    } catch (err) {
      setError(errText(err, 'Could not add the route.'));
    } finally {
      setSaving(false);
    }
  };

  // An operator admin has exactly one operator, so it is selected automatically.
  const vehicleOperatorId = vehicleForm.operatorId || (operators.length === 1 ? operators[0].id : '');
  const vehicleOperator = operators.find((o) => o.id === vehicleOperatorId);

  const setCoach = (index, name, value) =>
    setVehicleForm({
      ...vehicleForm,
      coaches: vehicleForm.coaches.map((c, i) => (i === index ? { ...c, [name]: value } : c)),
    });

  const handleCreateVehicle = async (e) => {
    e.preventDefault();
    if (!vehicleOperator) {
      setError('Please select an operator.');
      return;
    }
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const payload = {
        operator_id: vehicleOperator.id,
        registration_no: vehicleForm.registrationNo,
        type: vehicleOperator.type,
      };
      if (vehicleOperator.type === 'BUS') payload.bus_layout = vehicleForm.busLayout;
      if (vehicleOperator.type === 'TRAIN') {
        payload.coaches = vehicleForm.coaches.map((c) => ({ ...c, total_seats: Number(c.total_seats) }));
      }

      const res = await API.post('/admin/vehicles', payload);
      setMessage(`Vehicle ${res.data.registration_no} added. You can now schedule trips for it.`);
      setVehicleForm({ ...EMPTY_VEHICLE, operatorId: vehicleForm.operatorId });
      await load();
    } catch (err) {
      setError(errText(err, 'Could not add the vehicle.'));
    } finally {
      setSaving(false);
    }
  };

  const handleCreateOperator = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const res = await API.post('/admin/operators', operatorForm);
      setMessage(`Operator "${res.data.name}" added. Now add an admin for it below.`);
      setOperatorForm(EMPTY_OPERATOR);
      setAdminForm({ ...adminForm, operatorId: res.data.id });
      await load();
    } catch (err) {
      setError(errText(err, 'Could not add the operator.'));
    } finally {
      setSaving(false);
    }
  };

  const handleCreateOperatorAdmin = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    setMessage('');
    try {
      const res = await API.post('/admin/operator-admins', {
        operator_id: adminForm.operatorId,
        email: adminForm.email,
        full_name: adminForm.fullName,
        phone: adminForm.phone,
        password: adminForm.password,
      });
      setMessage(res.data.notice);
      setAdminForm({ ...EMPTY_ADMIN, operatorId: adminForm.operatorId });
      await load();
    } catch (err) {
      setError(errText(err, 'Could not add the operator admin.'));
    } finally {
      setSaving(false);
    }
  };

  const handleRemoveAdmin = async (admin) => {
    if (!window.confirm(`Remove ${admin.full_name} as admin of ${admin.operator_name}? They become a normal passenger.`)) return;
    setError('');
    setMessage('');
    try {
      await API.delete(`/admin/operator-admins/${admin.user_id}`);
      setMessage(`${admin.full_name} is no longer an operator admin.`);
      await load();
    } catch (err) {
      setError(errText(err, 'Could not remove the operator admin.'));
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
          <button type="button" className={activeTab === 'trips' ? 'on' : ''} onClick={() => switchTab('trips')}>
            <PlusCircle size={16} aria-hidden="true" /> Trips
          </button>
          <button type="button" className={activeTab === 'routes' ? 'on' : ''} onClick={() => switchTab('routes')}>
            <MapPin size={16} aria-hidden="true" /> Routes
          </button>
          <button type="button" className={activeTab === 'vehicles' ? 'on' : ''} onClick={() => switchTab('vehicles')}>
            <Bus size={16} aria-hidden="true" /> Vehicles
          </button>
          {isSuper && (
            <button type="button" className={activeTab === 'operators' ? 'on' : ''} onClick={() => switchTab('operators')}>
              <Building2 size={16} aria-hidden="true" /> Operators
            </button>
          )}
          <button type="button" className={activeTab === 'issue' ? 'on' : ''} onClick={() => switchTab('issue')}>
            <Ticket size={16} aria-hidden="true" /> Counter ticket
          </button>
          <button type="button" onClick={() => navigate('/profile')}>
            <Layers size={16} aria-hidden="true" /> Profile
          </button>
        </nav>

        <main className="ad-main">
          {error && <div className="ad-alert ad-alert--error" role="alert"><AlertCircle size={16} aria-hidden="true" /> {error}</div>}
          {message && <div className="ad-alert ad-alert--ok" role="status"><CheckCircle2 size={16} aria-hidden="true" /> {message}</div>}

          {activeTab === 'trips' && (
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
          )}

          {activeTab === 'routes' && (
            <>
              <section className="card">
                <h3 className="card-title">Add a route</h3>
                <p className="muted">
                  Routes are shared by all operators. Use the same spelling as the search page
                  (for example Chattogram, not Chittagong), and add the return route separately.
                </p>
                <form onSubmit={handleCreateRoute} className="form-grid ad-form">
                  <div className="form-field">
                    <label htmlFor="origin">From</label>
                    <input id="origin" value={routeForm.origin_city} maxLength={60} placeholder="e.g. Dhaka" required
                      onChange={(e) => setRouteForm({ ...routeForm, origin_city: e.target.value })} />
                  </div>
                  <div className="form-field">
                    <label htmlFor="destination">To</label>
                    <input id="destination" value={routeForm.destination_city} maxLength={60} placeholder="e.g. Rajshahi" required
                      onChange={(e) => setRouteForm({ ...routeForm, destination_city: e.target.value })} />
                  </div>
                  <div className="ad-actions">
                    <button type="submit" className="btn-primary" disabled={saving}>{saving ? 'Adding...' : 'Add route'}</button>
                  </div>
                </form>
              </section>

              <section className="card ad-list">
                <h3 className="card-title">All routes ({routes.length})</h3>
                <div className="ad-table-wrap">
                  <table className="ad-table">
                    <thead><tr><th>From</th><th>To</th></tr></thead>
                    <tbody>
                      {routes.map((r) => (
                        <tr key={r.id}><td>{r.origin_city}</td><td>{r.destination_city}</td></tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </section>
            </>
          )}

          {activeTab === 'vehicles' && (
            <>
              <section className="card">
                <h3 className="card-title">Add a vehicle</h3>
                <p className="muted">
                  A vehicle belongs to an operator and has the same type as that operator.
                  Its seats are created from the layout when you schedule a trip.
                </p>
                <form onSubmit={handleCreateVehicle} className="form-grid ad-form">
                  <div className="form-field">
                    <label htmlFor="operator">Operator</label>
                    <select id="operator" value={vehicleOperatorId} disabled={operators.length === 1} required
                      onChange={(e) => setVehicleForm({ ...vehicleForm, operatorId: e.target.value })}>
                      <option value="">Select an operator</option>
                      {operators.map((o) => (
                        <option key={o.id} value={o.id}>{o.name} ({o.type})</option>
                      ))}
                    </select>
                  </div>

                  <div className="form-field">
                    <label htmlFor="reg">Registration number</label>
                    <input id="reg" value={vehicleForm.registrationNo} maxLength={30} placeholder="e.g. DHAKA-METRO-BA-11-1234" required
                      onChange={(e) => setVehicleForm({ ...vehicleForm, registrationNo: e.target.value })} />
                  </div>

                  {vehicleOperator?.type === 'BUS' && (
                    <div className="form-field">
                      <label htmlFor="busLayout">Seat layout</label>
                      <select id="busLayout" value={vehicleForm.busLayout}
                        onChange={(e) => setVehicleForm({ ...vehicleForm, busLayout: e.target.value })}>
                        <option value="2x2">2 + 2 (40 seats)</option>
                        <option value="2x1">2 + 1 (30 seats)</option>
                      </select>
                    </div>
                  )}

                  {vehicleOperator?.type === 'LAUNCH' && (
                    <p className="muted ad-wide">Launches use the standard layout: 40 deck seats and 14 cabins (single, double and VIP).</p>
                  )}

                  {vehicleOperator?.type === 'TRAIN' && (
                    <div className="ad-wide">
                      <label className="ad-coach-title">Coaches</label>
                      {vehicleForm.coaches.map((c, i) => (
                        <div key={i} className="ad-coach">
                          <input aria-label="Coach name" placeholder="KA" maxLength={4} value={c.coach_name} required
                            onChange={(e) => setCoach(i, 'coach_name', e.target.value)} />
                          <select aria-label="Class" value={c.class} onChange={(e) => setCoach(i, 'class', e.target.value)}>
                            <option value="SNIGDHA">Snigdha</option>
                            <option value="SHOVON_CHAIR">Shovon chair</option>
                          </select>
                          <select aria-label="Layout" value={c.layout} onChange={(e) => setCoach(i, 'layout', e.target.value)}>
                            <option value="2x2">2 + 2</option>
                            <option value="2x3">2 + 3</option>
                          </select>
                          <input aria-label="Seats" type="number" min="1" max="120" value={c.total_seats} required
                            onChange={(e) => setCoach(i, 'total_seats', e.target.value)} />
                          <button type="button" className="btn-secondary" aria-label="Remove coach" disabled={vehicleForm.coaches.length === 1}
                            onClick={() => setVehicleForm({ ...vehicleForm, coaches: vehicleForm.coaches.filter((_, j) => j !== i) })}>
                            <Trash2 size={14} aria-hidden="true" />
                          </button>
                        </div>
                      ))}
                      <button type="button" className="btn-secondary" disabled={vehicleForm.coaches.length >= 12}
                        onClick={() => setVehicleForm({ ...vehicleForm, coaches: [...vehicleForm.coaches, { ...EMPTY_COACH }] })}>
                        Add coach
                      </button>
                    </div>
                  )}

                  <div className="ad-actions">
                    <button type="submit" className="btn-primary" disabled={saving || !vehicleOperator}>{saving ? 'Adding...' : 'Add vehicle'}</button>
                  </div>
                </form>
              </section>

              <section className="card ad-list">
                <h3 className="card-title">Vehicles ({vehicles.length})</h3>
                <div className="ad-table-wrap">
                  <table className="ad-table">
                    <thead><tr><th>Registration</th><th>Operator</th><th>Type</th><th>Layout</th></tr></thead>
                    <tbody>
                      {vehicles.map((v) => (
                        <tr key={v.id}>
                          <td>{v.registration_no}</td>
                          <td>{v.operator_name}</td>
                          <td>{v.type}</td>
                          <td>{layoutSummary(v)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </section>
            </>
          )}

          {activeTab === 'operators' && isSuper && (
            <>
              <section className="card">
                <h3 className="card-title">Add an operator</h3>
                <p className="muted">An operator is a company, like a bus line or a launch service. Pick its type once: all its vehicles are of that type.</p>
                <form onSubmit={handleCreateOperator} className="form-grid ad-form">
                  <div className="form-field">
                    <label htmlFor="opname">Operator name</label>
                    <input id="opname" value={operatorForm.name} maxLength={80} placeholder="e.g. Demo River Express" required
                      onChange={(e) => setOperatorForm({ ...operatorForm, name: e.target.value })} />
                  </div>
                  <div className="form-field">
                    <label htmlFor="optype">Type</label>
                    <select id="optype" value={operatorForm.type} onChange={(e) => setOperatorForm({ ...operatorForm, type: e.target.value })}>
                      <option value="BUS">Bus</option>
                      <option value="LAUNCH">Launch</option>
                      <option value="TRAIN">Train</option>
                    </select>
                  </div>
                  <div className="form-field">
                    <label htmlFor="opphone">Contact number</label>
                    <input id="opphone" type="tel" value={operatorForm.contact_phone} placeholder="01712345678" required
                      onChange={(e) => setOperatorForm({ ...operatorForm, contact_phone: e.target.value })} />
                  </div>
                  <div className="ad-actions">
                    <button type="submit" className="btn-primary" disabled={saving}>{saving ? 'Adding...' : 'Add operator'}</button>
                  </div>
                </form>
              </section>

              <section className="card">
                <h3 className="card-title">Add an operator admin</h3>
                <p className="muted">
                  If the email already belongs to a passenger account, that account is promoted and the
                  name, mobile and password below are ignored (they must sign in again). Otherwise a new
                  account is created with the password you set here.
                </p>
                <form onSubmit={handleCreateOperatorAdmin} className="form-grid ad-form">
                  <div className="form-field">
                    <label htmlFor="adop">Operator</label>
                    <select id="adop" value={adminForm.operatorId} required
                      onChange={(e) => setAdminForm({ ...adminForm, operatorId: e.target.value })}>
                      <option value="">Select an operator</option>
                      {operators.map((o) => <option key={o.id} value={o.id}>{o.name} ({o.type})</option>)}
                    </select>
                  </div>
                  <div className="form-field">
                    <label htmlFor="ademail">Email</label>
                    <input id="ademail" type="email" value={adminForm.email} placeholder="admin@example.com" required
                      onChange={(e) => setAdminForm({ ...adminForm, email: e.target.value })} />
                  </div>
                  <div className="form-field">
                    <label htmlFor="adname">Full name (new account)</label>
                    <input id="adname" value={adminForm.fullName} maxLength={100}
                      onChange={(e) => setAdminForm({ ...adminForm, fullName: e.target.value })} />
                  </div>
                  <div className="form-field">
                    <label htmlFor="adphone">Mobile (new account)</label>
                    <input id="adphone" type="tel" value={adminForm.phone} placeholder="01712345678"
                      onChange={(e) => setAdminForm({ ...adminForm, phone: e.target.value })} />
                  </div>
                  <div className="form-field">
                    <label htmlFor="adpass">Password (new account)</label>
                    <input id="adpass" type="password" value={adminForm.password} autoComplete="new-password" placeholder="8+ characters, a letter and a number"
                      onChange={(e) => setAdminForm({ ...adminForm, password: e.target.value })} />
                  </div>
                  <div className="ad-actions">
                    <button type="submit" className="btn-primary" disabled={saving}>{saving ? 'Saving...' : 'Add operator admin'}</button>
                  </div>
                </form>
              </section>

              <section className="card ad-list">
                <h3 className="card-title">Operators ({operators.length}) and their admins</h3>
                <div className="ad-table-wrap">
                  <table className="ad-table">
                    <thead><tr><th>Operator</th><th>Type</th><th>Admins</th></tr></thead>
                    <tbody>
                      {operators.map((o) => {
                        const admins = operatorAdmins.filter((a) => a.operator_id === o.id);
                        return (
                          <tr key={o.id}>
                            <td>{o.name}</td>
                            <td>{o.type}</td>
                            <td>
                              {admins.length === 0 && <span className="muted">No admin yet</span>}
                              {admins.map((a) => (
                                <div key={a.user_id} className="ad-adminrow">
                                  <span>{a.full_name} <span className="muted">({a.email})</span></span>
                                  <button type="button" className="btn-secondary" onClick={() => handleRemoveAdmin(a)}>Remove</button>
                                </div>
                              ))}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </section>
            </>
          )}

          {activeTab === 'issue' && (
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
