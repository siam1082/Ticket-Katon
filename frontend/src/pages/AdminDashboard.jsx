import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ShieldCheck, PlusCircle, Ticket, Layers, Calendar, Clock } from 'lucide-react';
import './AdminDashboard.css';

export default function AdminDashboard() {
    const { user, isAuthenticated, role } = useAuth();
    const navigate = useNavigate();

    const [activeTab, setActiveTab] = useState('schedules'); // 'schedules' or 'issue'

    // Trip schedule state
    const [vehicleType, setVehicleType] = useState('Bus');
    const [operatorName, setOperatorName] = useState('');
    const [fromCity, setFromCity] = useState('');
    const [toCity, setToCity] = useState('');
    const [departureTime, setDepartureTime] = useState('');
    const [basePrice, setBasePrice] = useState('');

    useEffect(() => {
        // Only operator admin can access this page
        if (!isAuthenticated || role !== 'admin') {
            navigate('/login');
        }
    }, [isAuthenticated, role, navigate]);

    if (!isAuthenticated || role !== 'admin') return null;

    const handleCreateSchedule = (e) => {
        e.preventDefault();
        // TODO: call backend API to create schedule
        alert(`New trip scheduled for ${operatorName} (${vehicleType}) from ${fromCity} to ${toCity}`);
        setOperatorName('');
        setFromCity('');
        setToCity('');
        setDepartureTime('');
        setBasePrice('');
    };

    return (
        <div className="govd-layout">
            <header className="govd-header">
                <div className="govd-brand">
                    <div className="govd-logo">TK</div>
                    <div>
                        <div className="govd-agency">TicketKaton</div>
                        <div className="govd-app">Management & Counter Portal</div>
                    </div>
                </div>
                <div className="govd-auth">
                    <span className="role-tag"><ShieldCheck size={14} /> OPERATOR ADMIN</span>
                    <div style={{ marginLeft: 12 }}>{user?.name}</div>
                </div>
            </header>

            <div className="govd-body">
                <aside className="govd-sidebar card">
                    <ul>
                        <li className={activeTab === 'schedules' ? 'active' : ''} onClick={() => setActiveTab('schedules')}>
                            <PlusCircle size={16} /> Create Trip Schedule
                        </li>
                        <li className={activeTab === 'issue' ? 'active' : ''} onClick={() => setActiveTab('issue')}>
                            <Ticket size={16} /> Issue Counter Ticket
                        </li>
                        <li onClick={() => navigate('/profile')}>
                            <Layers size={16} /> Profile
                        </li>
                    </ul>
                </aside>

                <main className="govd-main">
                    <div className="govd-breadcrumbs">
                        <span>Admin</span>
                        <span className="sep">/</span>
                        <strong>{activeTab === 'schedules' ? 'Create Schedule' : 'Issue Ticket'}</strong>
                    </div>

                    {activeTab === 'schedules' ? (
                        <section className="card">
                            <h3 className="card-title"><Layers size={18} /> Schedule a New Journey</h3>

                            <form onSubmit={handleCreateSchedule} className="form-grid">
                                <div className="form-field">
                                    <label>Vehicle Type</label>
                                    <select value={vehicleType} onChange={(e) => setVehicleType(e.target.value)}>
                                        <option value="Bus">Bus</option>
                                        <option value="Train">Train</option>
                                        <option value="Launch">Launch</option>
                                    </select>
                                </div>

                                <div className="form-field">
                                    <label>Operator / Transport Name</label>
                                    <input
                                        type="text"
                                        value={operatorName}
                                        onChange={(e) => setOperatorName(e.target.value)}
                                        placeholder="e.g. Green Line, Parabat"
                                        required
                                    />
                                </div>

                                <div className="form-field">
                                    <label>From (Departure)</label>
                                    <input
                                        type="text"
                                        value={fromCity}
                                        onChange={(e) => setFromCity(e.target.value)}
                                        placeholder="e.g. Dhaka"
                                        required
                                    />
                                </div>

                                <div className="form-field">
                                    <label>To (Destination)</label>
                                    <input
                                        type="text"
                                        value={toCity}
                                        onChange={(e) => setToCity(e.target.value)}
                                        placeholder="e.g. Cox's Bazar"
                                        required
                                    />
                                </div>

                                <div className="form-field">
                                    <label>Departure Date & Time</label>
                                    <input
                                        type="datetime-local"
                                        value={departureTime}
                                        onChange={(e) => setDepartureTime(e.target.value)}
                                        required
                                    />
                                </div>

                                <div className="form-field">
                                    <label>Base Fare (BDT)</label>
                                    <input
                                        type="number"
                                        value={basePrice}
                                        onChange={(e) => setBasePrice(e.target.value)}
                                        placeholder="e.g. 1200"
                                        required
                                    />
                                </div>

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
                            <h3 className="card-title"><Ticket size={18} /> Direct Counter Ticket Issue</h3>
                            <p className="muted">Counter ticket issuance interface connects with backend endpoint <code>/admin/tickets/issue</code>.</p>

                            <div style={{ marginTop: 12 }}>
                                <p className="muted">Quick action buttons for issuing tickets will go here.</p>
                                <div style={{ marginTop: 10, display: 'flex', gap: 8 }}>
                                    <button className="btn-primary">Issue Manual Ticket</button>
                                    <button className="btn-secondary">Scan QR / Lookup</button>
                                </div>
                            </div>
                        </section>
                    )}
                </main>
            </div>

        </div>
    );
}
