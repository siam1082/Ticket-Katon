import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import API from '../services/api';
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
    const fetch = async () => {
      setLoading(true);
      setError('');
      try {
        const res = await API.get('/schedules', { params: { type, from, to, date } });
        setResults(res.data?.data || res.data || []);
      } catch (err) {
        setError(err.response?.data?.message || 'Failed to load schedules');
      } finally {
        setLoading(false);
      }
    };
    fetch();
  }, [type, from, to, date]);

  const handleSelect = (schedule) => {
    navigate(`/seat-select?scheduleId=${schedule.id}`);
  };

  return (
    <div className="gov-search-results">
      <div className="card">
        <h2 className="card-title">Search Results</h2>
        <p className="muted">{type.toUpperCase()} — {from} ➔ {to} — {date}</p>

        {loading ? (
          <p className="muted">Loading schedules...</p>
        ) : error ? (
          <div className="error-box">{error}</div>
        ) : results.length === 0 ? (
          <p className="muted">No schedules found for selected route.</p>
        ) : (
          <div className="results-list">
            {results.map((s) => (
              <div key={s.id} className="result-item">
                <div>
                  <strong>{s.operator}</strong>
                  <div className="muted">{s.from} ➔ {s.to}</div>
                </div>
                <div className="result-meta">
                  <div>{new Date(s.departure).toLocaleString()}</div>
                  <div className="price">{s.base_fare} BDT</div>
                  <button className="btn-primary" onClick={() => handleSelect(s)}>Select Seats</button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
