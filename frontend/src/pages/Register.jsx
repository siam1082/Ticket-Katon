import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { UserPlus, Building2, ArrowRight } from 'lucide-react';
import API from '../services/api';
import './Register.css';

export default function Register() {
    const [formData, setFormData] = useState({
        name: '',
        email: '',
        phone: '',
        password: '',
        confirmPassword: ''
    });
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleChange = (e) => {
        setFormData({
            ...formData,
            [e.target.name]: e.target.value
        });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        if (formData.password !== formData.confirmPassword) {
            setError('Passwords do not match');
            return;
        }

        setLoading(true);
        try {
            const response = await API.post('/auth/register', {
                name: formData.name,
                email: formData.email,
                phone: formData.phone,
                password: formData.password,
            });

            const { user, token } = response.data;
            login(user, token, user?.role || 'user');
            navigate('/');
        } catch (err) {
            setError(err.response?.data?.message || 'Registration failed. Try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="gov-register-shell">
            <div className="gov-register-panel">
                <div className="gov-register-header">
                    <div className="gov-branding">
                        <div className="gov-brand-mark" aria-hidden="true">
                            <Building2 size={28} />
                        </div>
                        <div>
                            <span className="gov-eyebrow">Government Transport Portal</span>
                            <h1>Passenger Registration</h1>
                        </div>
                    </div>
                    <p className="gov-subtitle">Create your secure account to book tickets online</p>
                </div>

                <div className="gov-register-body">
                    <div className="gov-register-intro">
                        <UserPlus size={20} />
                        <span>New passenger account</span>
                    </div>

                    {error && <div className="gov-error-box">{error}</div>}

                    <form onSubmit={handleSubmit} className="gov-register-form">
                        <div className="gov-field-row">
                            <div className="gov-field">
                                <label htmlFor="name">Full Name</label>
                                <input
                                    id="name"
                                    type="text"
                                    name="name"
                                    value={formData.name}
                                    onChange={handleChange}
                                    placeholder="e.g. John Doe"
                                    required
                                />
                            </div>
                        </div>

                        <div className="gov-field-row two-col">
                            <div className="gov-field">
                                <label htmlFor="email">Email Address</label>
                                <input
                                    id="email"
                                    type="email"
                                    name="email"
                                    value={formData.email}
                                    onChange={handleChange}
                                    placeholder="name@example.com"
                                    required
                                />
                            </div>

                            <div className="gov-field">
                                <label htmlFor="phone">Mobile Number</label>
                                <input
                                    id="phone"
                                    type="tel"
                                    name="phone"
                                    value={formData.phone}
                                    onChange={handleChange}
                                    placeholder="017XXXXXXXX"
                                    required
                                />
                            </div>
                        </div>

                        <div className="gov-field-row two-col">
                            <div className="gov-field">
                                <label htmlFor="password">Password</label>
                                <input
                                    id="password"
                                    type="password"
                                    name="password"
                                    value={formData.password}
                                    onChange={handleChange}
                                    placeholder="••••••••"
                                    required
                                />
                            </div>

                            <div className="gov-field">
                                <label htmlFor="confirmPassword">Confirm Password</label>
                                <input
                                    id="confirmPassword"
                                    type="password"
                                    name="confirmPassword"
                                    value={formData.confirmPassword}
                                    onChange={handleChange}
                                    placeholder="••••••••"
                                    required
                                />
                            </div>
                        </div>

                        <button type="submit" className="gov-submit-btn" disabled={loading}>
                            {loading ? 'Creating Account...' : 'Create Account'} <ArrowRight size={16} />
                        </button>
                    </form>

                    <p className="gov-footer-text">
                        Already have an account?{' '}
                        <Link to="/login" className="gov-link">Login here</Link>
                    </p>
                </div>
            </div>
        </div>
    );
}
