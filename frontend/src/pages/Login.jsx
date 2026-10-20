import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { User, ShieldCheck, ArrowRight, Ticket } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import API from '../services/api';
import './Auth.css';

const STAFF_ROLES = ['OPERATOR_ADMIN', 'SUPER_ADMIN'];

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
      // One endpoint for everyone; the server decides the role.
      const { data } = await API.post('/auth/login', { email: email.trim(), password });
      const { user, token } = data;
      const isStaff = STAFF_ROLES.includes(user?.role);

      if (roleType === 'admin' && !isStaff) {
        setError('This account does not have admin access.');
        return;
      }

      login(user, token, isStaff ? 'admin' : 'user');
      navigate(isStaff ? '/admin' : '/');
    } catch (err) {
      setError(err.response?.data?.error || err.response?.data?.message || 'Invalid email or password.');
    } finally {
      setLoading(false);
    }
  };

  const isAdmin = roleType === 'admin';

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
              <h1>TicketKaton</h1>
            </div>
          </div>
          <p className="auth-sub">Sign in with a demo account to hold seats and view your tickets.</p>
        </div>

        <div className="auth-switch" role="tablist" aria-label="Login type">
          <button
            type="button"
            role="tab"
            aria-selected={!isAdmin}
            className={!isAdmin ? 'auth-switch-btn active' : 'auth-switch-btn'}
            onClick={() => setRoleType('user')}
          >
            <User size={16} aria-hidden="true" /> Passenger
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={isAdmin}
            className={isAdmin ? 'auth-switch-btn active' : 'auth-switch-btn'}
            onClick={() => setRoleType('admin')}
          >
            <ShieldCheck size={16} aria-hidden="true" /> Demo admin
          </button>
        </div>

        <div className="auth-body">
          <h2>{isAdmin ? 'Demo admin sign in' : 'Passenger sign in'}</h2>
          {isAdmin && <p className="auth-hint">Admin accounts are created for the presentation only.</p>}
          {error && <div className="auth-error" role="alert">{error}</div>}

          <form onSubmit={handleSubmit} className="auth-form">
            <div className="auth-field">
              <label htmlFor="email">Email</label>
              <input
                id="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder={isAdmin ? 'admin@demo.example' : 'passenger@example.com'}
                required
              />
            </div>

            <div className="auth-field">
              <label htmlFor="password">Password</label>
              <input
                id="password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>

            <button type="submit" className="auth-submit" disabled={loading}>
              {loading ? 'Signing in...' : 'Sign in'} <ArrowRight size={16} aria-hidden="true" />
            </button>
          </form>

          {!isAdmin && (
            <p className="auth-foot">
              Don&apos;t have an account? <Link to="/register" className="auth-link">Sign up</Link>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
