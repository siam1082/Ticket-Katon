import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { User, ShieldCheck, ArrowRight, Building2 } from 'lucide-react';
import API from '../services/api';
import './Login.css';

export default function Login() {
    const [roleType, setRoleType] = useState('user');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            const endpoint = roleType === 'admin' ? '/admin/auth/login' : '/auth/login';
            const response = await API.post(endpoint, {
                email,
                password
            });

            const { user, token } = response.data;
            login(user, token, roleType);

            if (roleType === 'admin') {
                navigate('/admin');
            } else {
                navigate('/');
            }
        } catch (err) {
            setError(err.response?.data?.message || 'Invalid credentials or login failed');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="gov-login-shell">
            <div className="gov-login-panel">
                <div className="gov-login-header">
                    <div className="gov-branding">
                        <div className="gov-brand-mark" aria-hidden="true">
                            <Building2 size={28} />
                        </div>
                        <div>
                            <span className="gov-eyebrow">Government Transport Portal</span>
                            <h1>TicketKaton</h1>
                        </div>
                    </div>
                    <p className="gov-subtitle">Secure access for passengers and operator administrators</p>
                </div>

                <div className="gov-role-switch" aria-label="Login role selector">
                    <button
                        type="button"
                        className={roleType === 'user' ? 'gov-role-btn active' : 'gov-role-btn'}
                        onClick={() => setRoleType('user')}
                    >
                        <User size={16} /> Passenger
                    </button>
                    <button
                        type="button"
                        className={roleType === 'admin' ? 'gov-role-btn active' : 'gov-role-btn'}
                        onClick={() => setRoleType('admin')}
                    >
                        <ShieldCheck size={16} /> Operator Admin
                    </button>
                </div>

                <div className="gov-login-body">
                    <h2>{roleType === 'admin' ? 'Operator Sign In' : 'Passenger Sign In'}</h2>
                    {error && <div className="gov-error-box">{error}</div>}

                    <form onSubmit={handleSubmit} className="gov-login-form">
                        <div className="gov-field">
                            <label htmlFor="email">
                                {roleType === 'admin' ? 'Operator Email / Username' : 'Email or Phone'}
                            </label>
                            <input
                                id="email"
                                type="text"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                placeholder={roleType === 'admin' ? 'admin@operator.com' : 'passenger@gmail.com'}
                                required
                            />
                        </div>

                        <div className="gov-field">
                            <label htmlFor="password">Password</label>
                            <input
                                id="password"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="••••••••"
                                required
                            />
                        </div>

                        <button type="submit" className="gov-submit-btn" disabled={loading}>
                            {loading ? 'Signing In...' : `Login as ${roleType === 'admin' ? 'Operator' : 'Passenger'}`} <ArrowRight size={16} />
                        </button>
                    </form>

                    {roleType === 'user' && (
                        <p className="gov-footer-text">
                            Don't have an account?{' '}
                            <Link to="/register" className="gov-link">Sign Up</Link>
                        </p>
                    )}
                </div>
            </div>
        </div>
    );
}
