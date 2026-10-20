import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import API from '../services/api';
import { Bus, Clock, MapPin, Ticket } from 'lucide-react';
import './SearchResults.css';

export default function SearchResults() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [results, setResults] = useState([]);
  const [error, setError] = useState('');

  const type = searchParams.get('type') || 'bus';
  const from = searchParams.get('from') || '';
  const to = searchParams.get('to') || '';
  const date = searchParams.get('date') || '';

  useEffect(() => {
    const fetchTrips = async () => {
      setLoading(true);
      setError('');
      try {
        const res = await API.get('/trips/search', {
          params: {
            origin: from,
            destination: to,
            date: date,
            type: type
          }
        });
        setResults(Array.isArray(res.data) ? res.data : (res.data?.data || []));
      } catch (err) {
        setError(err.response?.data?.error || 'Failed to load trip schedules');
      } finally {
        setLoading(false);
      }
    };
    fetchTrips();
  }, [type, from, to, date]);

  const handleSelect = (trip) => {
    navigate(`/seat-select?tripId=${trip.id}`);
  };

  return (
    <div className="gov-search-results" style={{ maxWidth: '900px', margin: '0 auto', padding: '20px' }}>
      <div className="card" style={{ background: '#fff', borderRadius: '12px', padding: '24px', boxShadow: '0 4px 12px rgba(0,0,0,0.05)' }}>
        <h2 style={{ fontSize: '20px', fontWeight: '700', marginBottom: '8px' }}>Available Schedules</h2>
        <p style={{ color: '#6b7280', marginBottom: '20px' }}>
          <strong>{type.toUpperCase()}</strong> • {from} ➔ {to} • {date}
        </p>

        {loading ? (
          <p style={{ textAlign: 'center', color: '#6b7280', padding: '40px' }}>Searching schedules from Ticket-Katon backend...</p>
        ) : error ? (
          <div style={{ background: '#fee2e2', color: '#b91c1c', padding: '12px', borderRadius: '8px' }}>{error}</div>
        ) : results.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <p style={{ color: '#4b5563', fontSize: '16px' }}>No trips found for this date & route.</p>
            <button className="btn-secondary" style={{ marginTop: '12px' }} onClick={() => navigate('/')}>Search Again</button>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
            {results.map((t) => (
              <div key={t.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', border: '1px solid #e5e7eb', padding: '16px 20px', borderRadius: '8px', background: '#fafafa' }}>
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontWeight: '700', fontSize: '18px' }}>
                    <Bus size={20} color="#15803d" />
                    <span>{t.operator_name}</span>
                    <span style={{ fontSize: '12px', background: '#e0e7ff', color: '#3730a3', padding: '2px 8px', borderRadius: '4px' }}>{t.registration_no}</span>
                  </div>
                  <div style={{ display: 'flex', gap: '16px', color: '#4b5563', marginTop: '8px', fontSize: '14px' }}>
                    <span><MapPin size={14} style={{ display: 'inline', verticalAlign: 'middle' }} /> {t.origin} ➔ {t.destination}</span>
                    <span><Clock size={14} style={{ display: 'inline', verticalAlign: 'middle' }} /> {new Date(t.departure_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                  </div>
                  <div style={{ marginTop: '6px', fontSize: '13px', color: '#16a34a', fontWeight: '600' }}>
                    <Ticket size={14} style={{ display: 'inline', verticalAlign: 'middle' }} /> {t.available_seats} seats available
                  </div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontSize: '22px', fontWeight: '800', color: '#111827', marginBottom: '8px' }}>{t.base_fare} BDT</div>
                  <button 
                    onClick={() => handleSelect(t)}
                    style={{ background: '#16a34a', color: '#fff', border: 'none', padding: '10px 18px', borderRadius: '6px', fontWeight: '600', cursor: 'pointer' }}
                  >
                    Select Seats
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}