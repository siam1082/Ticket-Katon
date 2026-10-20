import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { User, Mail, Shield, Ticket, Calendar, Clock, MapPin, CheckCircle2 } from 'lucide-react';
import API from '../services/api';
import './Profile.css';

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
                setBookings(Array.isArray(res.data) ? res.data : (res.data?.data || []));
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
        <div className="gov-profile-shell">
            <section className="gov-profile-header">
                <div className="gov-avatar">
                    <User size={36} color="#ffffff" />
                </div>
                <div className="gov-profile-meta">
                    <h2>{user?.name || 'Passenger'}</h2>
                    <div className="gov-meta-row">
                        <span className="gov-detail-pill"><Mail size={14} /> {user?.email || 'N/A'}</span>
                        <span className="gov-role-pill"><Shield size={12} /> {role?.toUpperCase() || 'USER'}</span>
                    </div>
                </div>
            </section>

            <section className="gov-profile-card">
                <div className="gov-section-header">
                    <Ticket size={20} />
                    <h3>My Booked Tickets</h3>
                </div>

                {loading ? (
                    <p className="gov-empty-text">Loading bookings...</p>
                ) : bookings.length === 0 ? (
                    <p className="gov-empty-text">No bookings found under this account.</p>
                ) : (
                    <div className="gov-ticket-grid">
                        {bookings.map((b) => (
                            <article key={b.id || b.pnr_code} className="gov-ticket-card">
                                <div className="gov-ticket-top">
                                    <div>
                                        <span className="gov-ticket-mode">{b.transport_type || 'BUS'}</span>
                                        <h4>{b.operator_name || 'Green Line Paribahan'}</h4>
                                        <span className="gov-ticket-id">PNR: <strong>{b.pnr_code}</strong></span>
                                    </div>
                                    <span className="gov-status-badge"><CheckCircle2 size={12} /> {b.status || 'CONFIRMED'}</span>
                                </div>

                                <div className="gov-route-row">
                                    <div>
                                        <small>Route</small>
                                        <p><MapPin size={12} /> {b.origin || 'Dhaka'} ➔ {b.destination || 'Chittagong'}</p>
                                    </div>
                                    <div>
                                        <small>Seats</small>
                                        <p>{Array.isArray(b.seats) ? b.seats.join(', ') : (b.seat_number || 'A1')}</p>
                                    </div>
                                    <div>
                                        <small>Total Fare</small>
                                        <p>{b.total_amount || b.fare || 800} BDT</p>
                                    </div>
                                </div>

                                <div className="gov-ticket-footer">
                                    <span><Calendar size={13} /> {b.departure_time ? new Date(b.departure_time).toLocaleDateString() : '2026-10-20'}</span>
                                    <span><Clock size={13} /> {b.departure_time ? new Date(b.departure_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '08:00 AM'}</span>
                                </div>
                            </article>
                        ))}
                    </div>
                )}
            </section>
        </div>
    );
}