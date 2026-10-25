import { useEffect, useMemo, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Bus, Train, Ship, Clock, MapPin, ArrowRight, Armchair, ChevronLeft, ChevronRight, Pencil } from 'lucide-react';
import API from '../services/api';
import PageLoader from '../components/PageLoader';
import { rememberTransportType } from '../hooks/useTransportType';
import './SearchResults.css';

const TYPES = {
  bus: { label: 'Bus', Icon: Bus },
  train: { label: 'Train', Icon: Train },
  launch: { label: 'Launch', Icon: Ship },
};

const DHAKA = 'Asia/Dhaka';

const formatTime = (v) => new Date(v).toLocaleTimeString('en-GB', { timeZone: DHAKA, hour: '2-digit', minute: '2-digit' });

function formatDay(isoDate) {
  if (!isoDate) return '';
  return new Date(`${isoDate}T12:00:00+06:00`).toLocaleDateString('en-GB', {
    timeZone: DHAKA, weekday: 'short', day: '2-digit', month: 'short', year: 'numeric',
  });
}

function formatDuration(start, end) {
  const minutes = Math.round((new Date(end) - new Date(start)) / 60000);
  if (!Number.isFinite(minutes) || minutes <= 0) return '';
  return `${Math.floor(minutes / 60)}h ${String(minutes % 60).padStart(2, '0')}m`;
}

function todayIso() {
  const d = new Date();
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
  return d.toISOString().slice(0, 10);
}

function shiftIso(isoDate, days) {
  const d = new Date(`${isoDate}T00:00:00Z`);
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}

const SORTS = {
  time: (a, b) => new Date(a.departure_time) - new Date(b.departure_time),
  price: (a, b) => a.base_fare - b.base_fare,
  seats: (a, b) => b.available_seats - a.available_seats,
};

function SeatsBadge({ available, total }) {
  if (available <= 0) return <span className="sr-seats sr-seats--out">Sold out</span>;
  if (available <= 5) return <span className="sr-seats sr-seats--low">Only {available} left</span>;
  return <span className="sr-seats">{available}{total ? ` of ${total}` : ''} seats available</span>;
}

export default function SearchResults() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const rawType = searchParams.get('type');
  const type = TYPES[rawType] ? rawType : 'bus';
  const from = searchParams.get('from') || '';
  const to = searchParams.get('to') || '';
  const date = searchParams.get('date') || '';

  const [attempt, setAttempt] = useState(0); // bumped by "Try again"
  const [sort, setSort] = useState('time');
  const [state, setState] = useState({ key: '', trips: [], error: '', at: 0 });

  // One key per search. While the stored result belongs to an older key, we are loading.
  const queryKey = `${type}|${from}|${to}|${date}|${attempt}`;
  const loading = state.key !== queryKey;

  useEffect(() => {
    rememberTransportType(type);
    let cancelled = false;

    (async () => {
      try {
        const res = await API.get('/trips/search', { params: { origin: from, destination: to, date, type } });
        const trips = Array.isArray(res.data) ? res.data : res.data?.data || [];
        if (!cancelled) setState({ key: queryKey, trips, error: '', at: Date.now() });
      } catch (err) {
        const error = err.response?.data?.error || 'Could not load the trips. Please try again.';
        if (!cancelled) setState({ key: queryKey, trips: [], error, at: Date.now() });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [queryKey, type, from, to, date]);

  const trips = useMemo(() => [...state.trips].sort(SORTS[sort]), [state.trips, sort]);

  const goToDate = (iso) => {
    const params = new URLSearchParams({ type, from, to, date: iso });
    navigate(`/search?${params.toString()}`, { replace: true });
  };

  const { Icon: TypeIcon, label: typeLabel } = TYPES[type];
  const canGoBack = date && date > todayIso();

  return (
    <div className="sr">
      <section className="sr-summary" aria-label="Your search">
        <div className="sr-summary-main">
          <span className="sr-type"><TypeIcon size={15} aria-hidden="true" /> {typeLabel}</span>
          <h2>
            {from} <ArrowRight size={18} aria-hidden="true" /> {to}
          </h2>
          <p>{formatDay(date)}</p>
        </div>
        <div className="sr-summary-actions">
          <button type="button" className="sr-ghost" onClick={() => navigate(`/?type=${type}`)}>
            <Pencil size={14} aria-hidden="true" /> Modify search
          </button>
        </div>
      </section>

      <div className="sr-days">
        <button type="button" className="sr-ghost" disabled={!canGoBack} onClick={() => goToDate(shiftIso(date, -1))}>
          <ChevronLeft size={16} aria-hidden="true" /> Previous day
        </button>
        <span>{formatDay(date)}</span>
        <button type="button" className="sr-ghost" disabled={!date} onClick={() => goToDate(shiftIso(date, 1))}>
          Next day <ChevronRight size={16} aria-hidden="true" />
        </button>
      </div>

      {loading ? (
        <PageLoader type={type} label="Searching trips…" />
      ) : state.error ? (
        <div className="sr-state sr-state--error" role="alert">
          <p>{state.error}</p>
          <button type="button" className="btn-primary" onClick={() => setAttempt((a) => a + 1)}>Try again</button>
        </div>
      ) : trips.length === 0 ? (
        <div className="sr-state">
          <h3>No {typeLabel.toLowerCase()} trips found</h3>
          <p>There are no trips from {from} to {to} on {formatDay(date)}. Try another day.</p>
          <div className="sr-state-actions">
            <button type="button" className="btn-primary" onClick={() => goToDate(shiftIso(date, 1))}>Check the next day</button>
            <button type="button" className="btn-secondary" onClick={() => navigate(`/?type=${type}`)}>Change route</button>
          </div>
        </div>
      ) : (
        <>
          <div className="sr-toolbar">
            <p>
              <strong>{trips.length}</strong> {trips.length === 1 ? 'trip' : 'trips'} found
            </p>
            <label>
              Sort by
              <select value={sort} onChange={(e) => setSort(e.target.value)}>
                <option value="time">Departure time</option>
                <option value="price">Lowest fare</option>
                <option value="seats">Most seats</option>
              </select>
            </label>
          </div>

          <ul className="sr-list">
            {trips.map((t) => {
              const kind = TYPES[String(t.transport_type || '').toLowerCase()] || TYPES[type];
              const Vehicle = kind.Icon;
              const departed = new Date(t.departure_time).getTime() <= state.at;
              const soldOut = t.available_seats <= 0;
              const duration = formatDuration(t.departure_time, t.arrival_time);

              return (
                <li key={t.id} className="sr-card">
                  <div className="sr-card-main">
                    <div className="sr-operator">
                      <span className="sr-vehicle"><Vehicle size={20} aria-hidden="true" /></span>
                      <div>
                        <strong>{t.operator_name}</strong>
                        <small>
                          {t.registration_no}
                          {t.category_name ? ` · ${t.category_name}` : ''}
                        </small>
                      </div>
                    </div>

                    <div className="sr-times">
                      <div>
                        <b>{formatTime(t.departure_time)}</b>
                        <small><MapPin size={12} aria-hidden="true" /> {t.origin}</small>
                      </div>
                      <div className="sr-line" aria-hidden="true">
                        <span><Clock size={12} /> {duration}</span>
                      </div>
                      <div>
                        <b>{formatTime(t.arrival_time)}</b>
                        <small><MapPin size={12} aria-hidden="true" /> {t.destination}</small>
                      </div>
                    </div>

                    <SeatsBadge available={t.available_seats} total={t.total_seats} />
                  </div>

                  <div className="sr-card-side">
                    <span className="sr-from">from</span>
                    <div className="sr-fare">{t.base_fare} <small>BDT</small></div>
                    <button
                      type="button"
                      className="btn-primary"
                      disabled={soldOut || departed}
                      onClick={() => navigate(`/seat-select?tripId=${t.id}`)}
                    >
                      <Armchair size={16} aria-hidden="true" />
                      {departed ? 'Departed' : soldOut ? 'Sold out' : 'Select seats'}
                    </button>
                  </div>
                </li>
              );
            })}
          </ul>
        </>
      )}
    </div>
  );
}
