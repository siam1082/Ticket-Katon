import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { User, Mail, Shield, Ticket, Calendar, Clock } from 'lucide-react';
import API from '../services/api';

export default function Profile() {
    const { user, isAuthenticated, role } = useAuth();
    const [bookings, setBookings] = useState([]);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        if (!isAuthenticated) {
            navigate('/login');
            return;
        }

        const fetchMyBookings = async () => {
            try {
                const res = await API.get('/bookings/my');
                setBookings(res.data?.data || res.data || []);
            } catch (err) {
                console.error('Failed to fetch user bookings', err);
            } finally {
                setLoading(false);
            }
        };

        fetchMyBookings();
    }, [isAuthenticated, navigate]);

    if (!isAuthenticated) return null;

    return (
        <div style={styles.container}>
            <div style={styles.profileHeader}>
                <div style={styles.avatar}>
                    <User size={36} color="#ffffff" />
                </div>
                <div>
                    <h2 style={styles.userName}>{user?.name || 'Passenger'}</h2>
                    <div style={styles.badgeRow}>
                        <span style={styles.emailText}><Mail size={14} /> {user?.email || 'N/A'}</span>
                        <span style={styles.roleBadge}><Shield size={12} /> {role?.toUpperCase() || 'USER'}</span>
                    </div>
                </div>
            </div>

            <div style={styles.bookingSection}>
                <div style={styles.sectionHeader}>
                    <Ticket size={20} color="#0066cc" />
                    <h3 style={styles.sectionTitle}>My Booked Tickets</h3>
                </div>

                {loading ? (
                    <p style={styles.emptyText}>Loading bookings...</p>
                ) : bookings.length === 0 ? (
                    <p style={styles.emptyText}>No bookings found under this account.</p>
                ) : (
                    <div style={styles.ticketGrid}>
                        {bookings.map((ticket) => (
                            <div key={ticket.id || ticket.booking_id} style={styles.ticketCard}>
                                <div style={styles.ticketTop}>
                                    <div>
                                        <span style={styles.ticketMode}>{ticket.mode || 'Trip'}</span>
                                        <h4 style={styles.operatorName}>{ticket.operator || 'Operator'}</h4>
                                        <span style={styles.ticketId}>ID: {ticket.id || ticket.booking_id}</span>
                                    </div>
                                    <span style={styles.statusBadge}>{ticket.status || 'Confirmed'}</span>
                                </div>

                                <div style={styles.routeRow}>
                                    <div>
                                        <small style={styles.routeLabel}>Route</small>
                                        <p style={styles.routeValue}>{ticket.from} ➔ {ticket.to}</p>
                                    </div>
                                    <div>
                                        <small style={styles.routeLabel}>Seats</small>
                                        <p style={styles.routeValue}>{Array.isArray(ticket.seats) ? ticket.seats.join(', ') : ticket.seats}</p>
                                    </div>
                                    <div>
                                        <small style={styles.routeLabel}>Fare</small>
                                        <p style={styles.routeValue}>{ticket.fare || ticket.total_amount} BDT</p>
                                    </div>
                                </div>

                                <div style={styles.ticketFooter}>
                                    <span><Calendar size={13} /> {ticket.date}</span>
                                    <span><Clock size={13} /> {ticket.time}</span>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}

const styles = {
    container: { maxWidth: '900px', margin: '0 auto', padding: '30px 16px' },
    profileHeader: { display: 'flex', alignItems: 'center', gap: '20px', backgroundColor: '#ffffff', padding: '24px', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.06)', marginBottom: '24px' },
    avatar: { width: '64px', height: '64px', borderRadius: '50%', backgroundColor: '#0066cc', display: 'flex', alignItems: 'center', justifyContent: 'center' },
    userName: { margin: 0, fontSize: '22px', color: '#111827' },
    badgeRow: { display: 'flex', alignItems: 'center', gap: '12px', marginTop: '6px' },
    emailText: { display: 'flex', alignItems: 'center', gap: '4px', fontSize: '14px', color: '#6b7280' },
    roleBadge: { display: 'flex', alignItems: 'center', gap: '4px', fontSize: '11px', fontWeight: 'bold', backgroundColor: '#e0f2fe', color: '#0369a1', padding: '2px 8px', borderRadius: '4px' },
    bookingSection: { backgroundColor: '#ffffff', padding: '24px', borderRadius: '8px', boxShadow: '0 2px 8px rgba(0,0,0,0.06)' },
    sectionHeader: { display: 'flex', alignItems: 'center', gap: '8px', borderBottom: '1px solid #f3f4f6', paddingBottom: '12px', marginBottom: '16px' },
    sectionTitle: { margin: 0, fontSize: '18px', color: '#111827' },
    emptyText: { color: '#6b7280', fontSize: '14px' },
    ticketGrid: { display: 'flex', flexDirection: 'column', gap: '14px' },
    ticketCard: { border: '1px solid #e5e7eb', borderRadius: '8px', padding: '16px', backgroundColor: '#fafafa' },
    ticketTop: { display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', borderBottom: '1px dashed #d1d5db', paddingBottom: '10px' },
    ticketMode: { fontSize: '12px', fontWeight: 'bold', color: '#0066cc', textTransform: 'uppercase' },
    operatorName: { margin: '2px 0', fontSize: '16px', color: '#111827' },
    ticketId: { fontSize: '12px', color: '#6b7280' },
    statusBadge: { backgroundColor: '#dcfce7', color: '#15803d', fontSize: '12px', fontWeight: 'bold', padding: '4px 8px', borderRadius: '4px' },
    routeRow: { display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px', padding: '12px 0' },
    routeLabel: { fontSize: '12px', color: '#6b7280' },
    routeValue: { margin: '2px 0 0', fontWeight: '600', fontSize: '14px', color: '#1f2937' },
    ticketFooter: { display: 'flex', gap: '16px', borderTop: '1px dashed #d1d5db', paddingTop: '10px', fontSize: '13px', color: '#4b5563' }
};