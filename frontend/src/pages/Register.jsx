import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { ArrowRight, Ticket, Check, Circle } from 'lucide-react';
import API from '../services/api';
import { EMAIL_RE, normalizePhone, passwordChecks, passwordStrength, STRENGTH_LABELS } from '../utils/validators';
import './Auth.css';

export default function Register() {
  const [form, setForm] = useState({ fullName: '', email: '', phone: '', password: '', confirmPassword: '' });
  const [touched, setTouched] = useState({});
  const [error, setError] = useState('');
  const [alreadyRegistered, setAlreadyRegistered] = useState(false);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const set = (e) => setForm({ ...form, [e.target.name]: e.target.value });
  const touch = (e) => setTouched({ ...touched, [e.target.name]: true });

  const fieldErrors = {
    fullName: form.fullName.trim().length < 2 ? 'Please enter your full name' : '',
    email: EMAIL_RE.test(form.email.trim()) ? '' : 'Please enter a valid email address',
    phone: normalizePhone(form.phone) ? '' : 'Enter a Bangladeshi mobile number, like 01712345678',
    password: passwordChecks(form.password).every((c) => c.ok) ? '' : 'Password does not meet the rules below',
    confirmPassword: form.confirmPassword === form.password ? '' : 'Passwords do not match',
  };
  const hasErrors = Object.values(fieldErrors).some(Boolean);
  const shown = (name) => (touched[name] ? fieldErrors[name] : '');
  const strength = passwordStrength(form.password);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setTouched({ fullName: true, email: true, phone: true, password: true, confirmPassword: true });
    setError('');
    setAlreadyRegistered(false);
    if (hasErrors) return;

    setLoading(true);
    try {
      const { data } = await API.post('/auth/register', {
        full_name: form.fullName.trim(),
        email: form.email.trim(),
        phone: normalizePhone(form.phone),
        password: form.password,
      });
      navigate('/verify-email', {
        state: { email: data.email, devCode: data.dev_code, emailSent: data.email_sent, message: data.message },
      });
    } catch (err) {
      if (err.response?.status === 409) setAlreadyRegistered(true);
      setError(err.response?.data?.error || 'Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-shell">
      <div className="auth-panel">
        <div className="auth-header">
          <div className="auth-brand">
            <div className="auth-mark" aria-hidden="true"><Ticket size={22} /></div>
            <div>
              <span className="auth-eyebrow">Academic demo</span>
              <h1>Create an account</h1>
            </div>
          </div>
          <p className="auth-sub">We will email you a 6-digit code to confirm your address.</p>
        </div>

        <div className="auth-body">
          {error && (
            <div className="auth-error" role="alert">
              {error}
              {alreadyRegistered && (
                <> <Link to="/verify-email" state={{ email: form.email.trim(), fromLogin: true }} className="auth-link">Verify your email</Link> or <Link to="/login" className="auth-link">sign in</Link>.</>
              )}
            </div>
          )}

          <form onSubmit={handleSubmit} className="auth-form" noValidate>
            <div className="auth-field">
              <label htmlFor="fullName">Full name</label>
              <input id="fullName" name="fullName" autoComplete="name" value={form.fullName} onChange={set} onBlur={touch}
                aria-invalid={!!shown('fullName')} placeholder="e.g. Rahim Uddin" />
              {shown('fullName') && <span className="auth-fielderr">{shown('fullName')}</span>}
            </div>

            <div className="auth-row two">
              <div className="auth-field">
                <label htmlFor="email">Email</label>
                <input id="email" name="email" type="email" autoComplete="email" value={form.email} onChange={set} onBlur={touch}
                  aria-invalid={!!shown('email')} placeholder="name@example.com" />
                {shown('email') && <span className="auth-fielderr">{shown('email')}</span>}
              </div>
              <div className="auth-field">
                <label htmlFor="phone">Mobile number</label>
                <input id="phone" name="phone" type="tel" autoComplete="tel" value={form.phone} onChange={set} onBlur={touch}
                  aria-invalid={!!shown('phone')} placeholder="01712345678" />
                {shown('phone') && <span className="auth-fielderr">{shown('phone')}</span>}
              </div>
            </div>

            <div className="auth-field">
              <label htmlFor="password">Password</label>
              <input id="password" name="password" type="password" autoComplete="new-password" value={form.password} onChange={set} onBlur={touch}
                aria-invalid={!!shown('password')} placeholder="At least 8 characters" />
              <div className="auth-meter" aria-hidden="true">
                {[1, 2, 3, 4].map((n) => <span key={n} className={n <= strength ? `on s${strength}` : ''} />)}
                <em>{STRENGTH_LABELS[strength]}</em>
              </div>
              <ul className="auth-rules">
                {passwordChecks(form.password).map((c) => (
                  <li key={c.label} className={c.ok ? 'ok' : ''}>
                    {c.ok ? <Check size={13} aria-hidden="true" /> : <Circle size={13} aria-hidden="true" />} {c.label}
                  </li>
                ))}
              </ul>
            </div>

            <div className="auth-field">
              <label htmlFor="confirmPassword">Confirm password</label>
              <input id="confirmPassword" name="confirmPassword" type="password" autoComplete="new-password" value={form.confirmPassword}
                onChange={set} onBlur={touch} aria-invalid={!!shown('confirmPassword')} placeholder="Repeat password" />
              {shown('confirmPassword') && <span className="auth-fielderr">{shown('confirmPassword')}</span>}
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
