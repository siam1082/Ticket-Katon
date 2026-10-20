import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { ArrowRight, Ticket } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import API from '../services/api';
import './Auth.css';

const STAFF_ROLES = ['OPERATOR_ADMIN', 'SUPER_ADMIN'];

export default function Register() {
  const [formData, setFormData] = useState({
    fullName: '',
    email: '',
    phone: '',
    password: '',
    confirmPassword: '',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (formData.password.length < 6) {
      setError('Password must be at least 6 characters.');
      return;
    }
    if (formData.password !== formData.confirmPassword) {
      setError('Passwords do not match.');
      return;
    }

    setLoading(true);
    try {
      // No role is sent: the server always creates a passenger account.
      const { data } = await API.post('/auth/register', {
        full_name: formData.fullName.trim(),
        email: formData.email.trim(),
        phone: formData.phone.trim(),
        password: formData.password,
      });

      const { user, token } = data;
      if (token) {
        login(user, token, STAFF_ROLES.includes(user?.role) ? 'admin' : 'user');
        navigate('/');
      } else {
        navigate('/login');
      }
    } catch (err) {
      setError(err.response?.data?.error || err.response?.data?.message || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-shell">
      <div className="auth-panel">
        <div className="auth-header">
          <div className="auth-brand">
            <div className="auth-mark" aria-hidden="true">
              <Ticket size={22} />
            </div>
            <div>
              <span className="auth-eyebrow">Academic demo</span>
              <h1>Create a demo account</h1>
            </div>
          </div>
          <p className="auth-sub">Use any email. Do not enter real personal or payment details.</p>
        </div>

        <div className="auth-body">
          {error && <div className="auth-error" role="alert">{error}</div>}

          <form onSubmit={handleSubmit} className="auth-form">
            <div className="auth-field">
              <label htmlFor="fullName">Full name</label>
              <input id="fullName" name="fullName" type="text" autoComplete="name" value={formData.fullName} onChange={handleChange} placeholder="e.g. Rahim Uddin" required />
            </div>

            <div className="auth-row two">
              <div className="auth-field">
                <label htmlFor="email">Email</label>
                <input id="email" name="email" type="email" autoComplete="email" value={formData.email} onChange={handleChange} placeholder="name@example.com" required />
              </div>
              <div className="auth-field">
                <label htmlFor="phone">Mobile number</label>
                <input id="phone" name="phone" type="tel" autoComplete="tel" value={formData.phone} onChange={handleChange} placeholder="017XXXXXXXX" required />
              </div>
            </div>

            <div className="auth-row two">
              <div className="auth-field">
                <label htmlFor="password">Password</label>
                <input id="password" name="password" type="password" autoComplete="new-password" value={formData.password} onChange={handleChange} placeholder="At least 6 characters" required />
              </div>
              <div className="auth-field">
                <label htmlFor="confirmPassword">Confirm password</label>
                <input id="confirmPassword" name="confirmPassword" type="password" autoComplete="new-password" value={formData.confirmPassword} onChange={handleChange} placeholder="Repeat password" required />
              </div>
            </div>

            <button type="submit" className="auth-submit" disabled={loading}>
              {loading ? 'Creating account...' : 'Create account'} <ArrowRight size={16} aria-hidden="true" />
            </button>
          </form>

          <p className="auth-foot">
            Already have an account? <Link to="/login" className="auth-link">Sign in</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
