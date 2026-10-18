import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ShieldCheck, PlusCircle, Ticket, Layers, Calendar, Clock, Home } from 'lucide-react';
import './AdminDashboard.css';

export default function AdminDashboard() {
  const { user, isAuthenticated, role } = useAuth();
  const navigate = useNavigate();

  const [activeTab, setActiveTab] = useState('schedules');

  // Trip schedule state
  const [vehicleType, setVehicleType] = useState('Bus');
  const [operatorName, setOperatorName] = useState('');
  const [fromCity, setFromCity] = useState('');
  const [toCity, setToCity] = useState('');
  const [departureTime, setDepartureTime] = useState('');
  const [basePrice, setBasePrice] = useState('');

  useEffect(() => {
    // Only admin can access this page
    if (!isAuthenticated || role !== 'admin') {
      navigate('/login');
    }
  }, [isAuthenticated, role, navigate]);

  if (!isAuthenticated || role !== 'admin') return null;

  const handleCreateSchedule = (e) => {
    e.preventDefault();
    // TODO: replace alert with real API call
    alert(`New trip scheduled for ${operatorName} (${vehicleType}) from ${fromCity} to ${toCity}`);
    setOperatorName('');
    setFromCity('');
    setToCity('');
    setDepartureTime('');
    setBasePrice('');
  };

  return (
    <div className="govd-layout">
      <header className="govd-header" role="banner">
        <div className="govd-brand">
          <div className="govd-logo" aria-hidden="true">LOGO</div>
          <div>
            <div className="govd-agency">Department of Transport</div>
            <div className="govd-app">Ticket Management — Operator Admin Console</div>
          </div>
        </div>
        <div className="govd-auth">
          <ShieldCheck size={16} /> <span className="govd-auth-text">Operator Admin</span>
          <span className="govd-user"> {user?.name || 'Operator'}</span>
        </div>
      </header>

      <div className="govd-body">
        <nav className="govd-sidebar" aria-label="Admin navigation">
          <ul>
            <li className={activeTab === 'schedules' ? 'active' : ''} onClick={() => setActiveTab('schedules')}>
              <Layers size={16} /> <span>Create Trip Schedule</span>
            </li>
            <li className={activeTab === 'issue' ? 'active' : ''} onClick={() => setActiveTab('issue')}>
              <Ticket size={16} /> <span>Issue Counter Ticket</span>
            </li>
            <li onClick={() => navigate('/admin/reports')}>
              <Calendar size={16} /> <span>Reports</span>
            </li>
            <li onClick={() => navigate('/admin/operations')}>
              <Clock size={16} /> <span>Operations</span>
            </li>
          </ul>
        </nav>

        <main className="govd-main" role="main">
          <div className="govd-breadcrumbs" aria-hidden="true">
            <Home size={14} /> <span>Home</span> <span className="sep">/</span> <span>Admin</span>
            <span className="sep">/</span> <strong>{activeTab === 'schedules' ? 'Create Trip Schedule' : 'Issue Ticket'}</strong>
          </div>

          {activeTab === 'schedules' ? (
            <section className="card">
              <h2 className="card-title"><PlusCircle size={18} /> Schedule a New Journey</h2>
              <form onSubmit={handleCreateSchedule} className="form-grid" aria-labelledby="schedule-form">
                <fieldset className="form-field">
                  <label htmlFor="vehicleType">Vehicle Type</label>
                  <select id="vehicleType" value={vehicleType} onChange={(e) => setVehicleType(e.target.value)}>
                    <option value="Bus">Bus</option>
                    <option value="Train">Train</option>
                    <option value="Launch">Launch</option>
                  </select>
                </fieldset>

                <fieldset className="form-field">
                  <label htmlFor="operatorName">Operator / Transport Name</label>
                  <input
                    id="operatorName"
                    type="text"
                    value={operatorName}
                    onChange={(e) => setOperatorName(e.target.value)}
                    placeholder="e.g. Green Line"
                    required
                  />
                </fieldset>

                <fieldset className="form-field">
                  <label htmlFor="fromCity">From (Departure)</label>
                  <input id="fromCity" type="text" value={fromCity} onChange={(e) => setFromCity(e.target.value)} placeholder="e.g. Dhaka" required />
                </fieldset>

                <fieldset className="form-field">
                  <label htmlFor="toCity">To (Destination)</label>
                  <input id="toCity" type="text" value={toCity} onChange={(e) => setToCity(e.target.value)} placeholder="e.g. Cox's Bazar" required />
                </fieldset>

                <fieldset className="form-field">
                  <label htmlFor="departureTime">Departure Date & Time</label>
                  <input id="departureTime" type="datetime-local" value={departureTime} onChange={(e) => setDepartureTime(e.target.value)} required />
                </fieldset>

                <fieldset className="form-field">
                  <label htmlFor="basePrice">Base Fare (BDT)</label>
                  <input id="basePrice" type="number" value={basePrice} onChange={(e) => setBasePrice(e.target.value)} placeholder="e.g. 1200" required />
                </fieldset>

                <div className="form-actions">
                  <button type="submit" className="btn-primary">Publish Trip Schedule</button>
                  <button type="button" className="btn-secondary" onClick={() => {
                    setOperatorName(''); setFromCity(''); setToCity(''); setDepartureTime(''); setBasePrice('');
                  }}>Reset</button>
                </div>
              </form>
            </section>
          ) : (
            <section className="card">
              <h2 className="card-title"><Ticket size={18} /> Direct Counter Ticket Issue</h2>
              <p className="muted">
                Use the counter interface to issue walk-in tickets. This connects to the backend endpoint <code>/admin/tickets/issue</code>.
              </p>
              <div className="placeholder-panel">
                <p className="placeholder-text">Counter ticket issuance UI will appear here. (Search by passenger name or mobile, select schedule, and issue ticket.)</p>
              </div>
            </section>
          )}
        </main>
      </div>

      <footer className="govd-footer" role="contentinfo">
        <div>© {new Date().getFullYear()} Department of Transport — Ticketing System</div>
      </footer>
    </div>
  );
}
