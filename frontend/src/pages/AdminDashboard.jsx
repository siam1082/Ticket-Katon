import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ShieldCheck, PlusCircle, Ticket, Layers, Calendar, Clock } from 'lucide-react';

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
        alert(`New trip scheduled for ${operatorName} (${vehicleType}) from ${fromCity} to ${toCity}`);
        setOperatorName('');
        setFromCity('');
        setToCity('');
        setDepartureTime('');
        setBasePrice('');
    };

    return (
        <div style={styles.container}>
            <div style={styles.topBanner}>
                <div>
                    <span style={styles.roleTag}><ShieldCheck size={14} /> OPERATOR ADMIN CONSOLE</span>
                    <h2 style={styles.title}>Management & Counter Portal</h2>
                    <p style={styles.subtitle}>Welcome back, {user?.name || 'Operator'}</p>
                </div>
            </div>

            <div style={styles.tabNav}>
                <button
                    style={{
                        ...styles.tabBtn,
                        borderBottom: activeTab === 'schedules' ? '3px solid #0066cc' : 'none',
                        color: activeTab === 'schedules' ? '#0066cc' : '#4b5563',
                        fontWeight: activeTab === 'schedules' ? '600' : '500'
                    }}
                    onClick={() => setActiveTab('schedules')}
                >
                    <PlusCircle size={16} /> Create Trip Schedule
                </button>
                <button
                    style={{
                        ...styles.tabBtn,
                        borderBottom: activeTab === 'issue' ? '3px solid #0066cc' : 'none',
                        color: activeTab === 'issue' ? '#0066cc' : '#4b5563',
                        fontWeight: activeTab === 'issue' ? '600' : '500'
                    }}
                    onClick={() => setActiveTab('issue')}
                >
                    <Ticket size={16} /> Issue Counter Ticket
                </button>
            </div>

            {activeTab === 'schedules' ? (
                <div style={styles.card}>
                    <h3 style={styles.cardHeading}><Layers size={18} /> Schedule a New Journey</h3>
                    <form onSubmit={handleCreateSchedule} style={styles.formGrid}>
                        <div style={styles.field}>
                            <label style={styles.label}>Vehicle Type</label>
                            <select
                                value={vehicleType}
                                onChange={(e) => setVehicleType(e.target.value)}
                                style={styles.select}
                            >
                                <option value="Bus">Bus</option>
                                <option value="Train">Train</option>
                                <option value="Launch">Launch</option>
                            </select>
                        </div>

                        <div style={styles.field}>
                            <label style={styles.label}>Operator / Transport Name</label>
                            <input
                                type="text"
                                value={operatorName}
                                onChange={(e) => setOperatorName(e.target.value)}
                                placeholder="e.g. Green Line, Parabat"
                                style={styles.input}
                                required
                            />
                        </div>

                        <div style={styles.field}>
                            <label style={styles.label}>From (Departure)</label>
                            <input
                                type="text"
                                value={fromCity}
                                onChange={(e) => setFromCity(e.target.value)}
                                placeholder="e.g. Dhaka"
                                style={styles.input}
                                required
                            />
                        </div>

                        <div style={styles.field}>
                            <label style={styles.label}>To (Destination)</label>
                            <input
                                type="text"
                                value={toCity}
                                onChange={(e) => setToCity(e.target.value)}
                                placeholder="e.g. Cox's Bazar"
                                style={styles.input}
                                required
                            />
                        </div>

                        <div style={styles.field}>
                            <label style={styles.label}>Departure Date & Time</label>
                            <input
                                type="datetime-local"
                                value={departureTime}
                                onChange={(e) => setDepartureTime(e.target.value)}
                                style={styles.input}
                                required
                            />
                        </div>

                        <div style={styles.field}>
                            <label style={styles.label}>Base Fare (BDT)</label>
                            <input
                                type="number"
                                value={basePrice}
                                onChange={(e) => setBasePrice(e.target.value)}
                                placeholder="e.g. 1200"
                                style={styles.input}
                                required
                            />
                        </div>

                        <div style={{ gridColumn: '1 / -1', marginTop: '10px' }}>
                            <button type="submit" style={styles.submitBtn}>
                                Publish Trip Schedule
                            </button>
                        </div>
                    </form>
                </div>
            ) : (
                <div style={styles.card}>
                    <h3 style={styles.cardHeading}><Ticket size={18} /> Direct Counter Ticket Issue</h3>
                    <p style={styles.placeholderText}>
                        Counter ticket issuance interface connects with backend endpoint <code>/admin/tickets/issue</code>.
                    </p>
                </div>
            )}
        </div>
    );
}

const styles = {
    container: { maxWidth: '1000px', margin: '0 auto', padding: '30px 16px' },
    topBanner: { backgroundColor: '#ffffff', padding: '24px', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.06)', marginBottom: '24px' },
    roleTag: { display: 'inline-flex', alignItems: 'center', gap: '4px', fontSize: '11px', fontWeight: 'bold', backgroundColor: '#e0f2fe', color: '#0369a1', padding: '2px 8px', borderRadius: '4px', marginBottom: '8px' },
    title: { margin: 0, fontSize: '22px', color: '#111827' },
    subtitle: { margin: '4px 0 0', fontSize: '14px', color: '#6b7280' },
    tabNav: { display: 'flex', gap: '20px', borderBottom: '1px solid #e5e7eb', marginBottom: '20px' },
    tabBtn: { display: 'flex', alignItems: 'center', gap: '6px', background: 'none', border: 'none', padding: '8px 4px', cursor: 'pointer', fontSize: '15px' },
    card: { backgroundColor: '#ffffff', padding: '24px', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.06)' },
    cardHeading: { display: 'flex', alignItems: 'center', gap: '8px', margin: '0 0 20px', fontSize: '17px', color: '#111827' },
    formGrid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '16px' },
    field: { display: 'flex', flexDirection: 'column', gap: '6px' },
    label: { fontSize: '13px', fontWeight: '500', color: '#374151' },
    input: { padding: '10px 12px', border: '1px solid #d1d5db', borderRadius: '6px', fontSize: '14px', outline: 'none' },
    select: { padding: '10px 12px', border: '1px solid #d1d5db', borderRadius: '6px', fontSize: '14px', outline: 'none', backgroundColor: '#fff' },
    submitBtn: { padding: '12px 24px', backgroundColor: '#0066cc', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '600' },
    placeholderText: { color: '#6b7280', fontSize: '14px' }
};