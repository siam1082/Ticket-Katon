import { useEffect, useState } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { MailCheck, ArrowRight, Info } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import API from '../services/api';
import { EMAIL_RE } from '../utils/validators';
import './Auth.css';

const RESEND_SECONDS = 60;
const STAFF_ROLES = ['OPERATOR_ADMIN', 'SUPER_ADMIN'];

export default function VerifyEmail() {
  const { state } = useLocation();
  const navigate = useNavigate();
  const { login } = useAuth();

  const [email, setEmail] = useState(state?.email || '');
  const [code, setCode] = useState('');
  const [devCode, setDevCode] = useState(state?.devCode || '');
  const [message, setMessage] = useState(state?.message || '');
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  // Came straight from sign-up: a code was just sent, so wait before offering another one.
  const [cooldown, setCooldown] = useState(state?.emailSent === undefined ? 0 : RESEND_SECONDS);

  useEffect(() => {
    if (cooldown <= 0) return undefined;
    const id = setTimeout(() => setCooldown((c) => c - 1), 1000);
    return () => clearTimeout(id);
  }, [cooldown]);

  const handleVerify = async (e) => {
    e.preventDefault();
    setError('');
    if (!EMAIL_RE.test(email.trim())) {
      setError('Please enter the email you signed up with.');
      return;
    }
    if (!/^\d{6}$/.test(code)) {
      setError('The code has 6 digits.');
      return;
    }

    setBusy(true);
    try {
      const { data } = await API.post('/auth/verify-email', { email: email.trim(), code });
      const isStaff = STAFF_ROLES.includes(data.user?.role);
      login(data.user, data.token, isStaff ? 'admin' : 'user');
      navigate('/');
    } catch (err) {
      setError(err.response?.data?.error || 'Could not verify the code. Please try again.');
    } finally {
      setBusy(false);
    }
  };

  const handleResend = async () => {
    setError('');
    if (!EMAIL_RE.test(email.trim())) {
      setError('Please enter your email first.');
      return;
    }
    setBusy(true);
    try {
      const { data } = await API.post('/auth/resend-code', { email: email.trim() });
      setMessage(data.message);
      setDevCode(data.dev_code || '');
      setCooldown(RESEND_SECONDS);
    } catch (err) {
      setError(err.response?.data?.error || 'Could not send a new code.');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="auth-shell">
      <div className="auth-panel">
        <div className="auth-header">
          <div className="auth-brand">
            <div className="auth-mark" aria-hidden="true"><MailCheck size={22} /></div>
            <div>
              <span className="auth-eyebrow">Almost there</span>
              <h1>Check your email</h1>
            </div>
          </div>
          <p className="auth-sub">
            Enter the 6-digit code we sent{email ? <> to <strong>{email}</strong></> : ''}. It is valid for 10 minutes.
          </p>
        </div>

        <div className="auth-body">
          {message && <div className="auth-info" role="status">{message}</div>}

          {devCode && (
            <div className="auth-dev" role="note">
              <Info size={16} aria-hidden="true" />
              <span>
                <strong>Demo mode:</strong> this server has no email provider set up, so no email was sent.
                Your code is <code>{devCode}</code>.
              </span>
            </div>
          )}

          {error && <div className="auth-error" role="alert">{error}</div>}

          <form onSubmit={handleVerify} className="auth-form" noValidate>
            {!state?.email && (
              <div className="auth-field">
                <label htmlFor="email">Email</label>
                <input id="email" type="email" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="name@example.com" />
              </div>
            )}

            <div className="auth-field">
              <label htmlFor="code">Verification code</label>
              <input id="code" className="auth-code" inputMode="numeric" autoComplete="one-time-code" maxLength={6} value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))} placeholder="123456" autoFocus />
            </div>

            <button type="submit" className="auth-submit" disabled={busy}>
              {busy ? 'Checking...' : 'Verify and continue'} <ArrowRight size={16} aria-hidden="true" />
            </button>
          </form>

          <p className="auth-foot">
            Did not get it?{' '}
            <button type="button" className="auth-linkbtn" onClick={handleResend} disabled={busy || cooldown > 0}>
              {cooldown > 0 ? `Send a new code in ${cooldown}s` : 'Send a new code'}
            </button>
          </p>
          <p className="auth-foot"><Link to="/login" className="auth-link">Back to sign in</Link></p>
        </div>
      </div>
    </div>
  );
}
