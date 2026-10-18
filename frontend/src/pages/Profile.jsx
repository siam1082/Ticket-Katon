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
                        {bookings.map((ticket) => (
                            <article key={ticket.id || ticket.booking_id} className="gov-ticket-card">
                                <div className="gov-ticket-top">
                                    <div>
                                        <span className="gov-ticket-mode">{ticket.mode || 'Trip'}</span>
                                        <h4>{ticket.operator || 'Operator'}</h4>
                                        <span className="gov-ticket-id">ID: {ticket.id || ticket.booking_id}</span>
                                    </div>
                                    <span className="gov-status-badge"><CheckCircle2 size={12} /> {ticket.status || 'Confirmed'}</span>
                                </div>

                                <div className="gov-route-row">
                                    <div>
                                        <small>Route</small>
                                        <p><MapPin size={12} /> {ticket.from} ➔ {ticket.to}</p>
                                    </div>
                                    <div>
                                        <small>Seats</small>
                                        <p>{Array.isArray(ticket.seats) ? ticket.seats.join(', ') : ticket.seats}</p>
                                    </div>
                                    <div>
                                        <small>Fare</small>
                                        <p>{ticket.fare || ticket.total_amount} BDT</p>
                                    </div>
                                </div>

                                <div className="gov-ticket-footer">
                                    <span><Calendar size={13} /> {ticket.date}</span>
                                    <span><Clock size={13} /> {ticket.time}</span>
                                </div>
                            </article>
                        ))}
                    </div>
                )}
            </section>
        </div>
    );
}
