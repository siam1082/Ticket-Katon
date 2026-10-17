import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { User, ShieldCheck } from 'lucide-react';
import API from '../services/api';

export default function Login() {
    const [roleType, setRoleType] = useState('user'); // 'user' or 'admin'
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
        <div style={styles.container}>
            <div style={styles.card}>
                <h2 style={styles.title}>Sign In to TicketKaton</h2>

                <div style={styles.tabContainer}>
                    <button
                        type="button"
                        style={{
                            ...styles.tabBtn,
                            backgroundColor: roleType === 'user' ? '#0066cc' : '#f3f4f6',
                            color: roleType === 'user' ? '#ffffff' : '#374151'
                        }}
                        onClick={() => setRoleType('user')}
                    >
                        <User size={16} /> Passenger
                    </button>
                    <button
                        type="button"
                        style={{
                            ...styles.tabBtn,
                            backgroundColor: roleType === 'admin' ? '#0066cc' : '#f3f4f6',
                            color: roleType === 'admin' ? '#ffffff' : '#374151'
                        }}
                        onClick={() => setRoleType('admin')}
                    >
                        <ShieldCheck size={16} /> Operator Admin
                    </button>
                </div>

                {error && <div style={styles.errorBox}>{error}</div>}

                <form onSubmit={handleSubmit} style={styles.form}>
                    <div style={styles.field}>
                        <label style={styles.label}>
                            {roleType === 'admin' ? 'Operator Email / Username' : 'Email or Phone'}
                        </label>
                        <input
                            type="text"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            placeholder={roleType === 'admin' ? 'admin@operator.com' : 'passenger@gmail.com'}
                            style={styles.input}
                            required
                        />
                    </div>

                    <div style={styles.field}>
                        <label style={styles.label}>Password</label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="••••••••"
                            style={styles.input}
                            required
                        />
                    </div>

                    <button type="submit" style={styles.submitBtn} disabled={loading}>
                        {loading ? 'Signing In...' : `Login as ${roleType === 'admin' ? 'Operator' : 'Passenger'}`}
                    </button>
                </form>

                {roleType === 'user' && (
                    <p style={styles.footerText}>
                        Don't have an account?{' '}
                        <Link to="/register" style={styles.link}>
                            Sign Up
                        </Link>
                    </p>
                )}
            </div>
        </div>
    );
}

const styles = {
    container: { display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '75vh' },
    card: { backgroundColor: '#ffffff', padding: '32px', borderRadius: '8px', boxShadow: '0 4px 12px rgba(0,0,0,0.08)', width: '100%', maxWidth: '400px' },
    title: { textAlign: 'center', marginBottom: '20px', fontSize: '22px', color: '#111827' },
    tabContainer: { display: 'flex', gap: '8px', marginBottom: '20px' },
    tabBtn: { flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px', padding: '10px', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '500' },
    form: { display: 'flex', flexDirection: 'column', gap: '16px' },
    field: { display: 'flex', flexDirection: 'column', gap: '6px' },
    label: { fontSize: '14px', fontWeight: '500', color: '#374151' },
    input: { padding: '10px 12px', border: '1px solid #d1d5db', borderRadius: '6px', fontSize: '14px', outline: 'none' },
    submitBtn: { padding: '12px', backgroundColor: '#0066cc', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '600', marginTop: '8px' },
    errorBox: { padding: '10px', backgroundColor: '#fee2e2', color: '#b91c1c', borderRadius: '6px', fontSize: '13px', marginBottom: '12px' },
    footerText: { textAlign: 'center', marginTop: '16px', fontSize: '14px', color: '#6b7280' },
    link: { color: '#0066cc', textDecoration: 'none', fontWeight: '600' }
};