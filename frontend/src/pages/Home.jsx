import { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Bus, Train, Ship, Search, Calendar, MapPin, Armchair, CreditCard, Ticket } from 'lucide-react';
import CityInput from '../components/CityInput';
import API from '../services/api';
import './Home.css';

const TYPES = {
  bus: { label: 'Bus', Icon: Bus, tagline: 'Coach seat maps with aisle layout and AC / non-AC classes.' },
  train: { label: 'Train', Icon: Train, tagline: 'Pick a coach and class, then choose your seat or berth.' },
  launch: { label: 'Launch', Icon: Ship, tagline: 'Deck seats and single, double or VIP cabins on river routes.' },
};

const STEPS = [
  { Icon: Search, title: 'Search', text: 'Choose route and date to see demo trips.' },
  { Icon: Armchair, title: 'Pick seats', text: 'Seats are held for you for 5 minutes.' },
  { Icon: CreditCard, title: 'Demo payment', text: 'Success, fail or cancel. No real money moves.' },
  { Icon: Ticket, title: 'Get ticket', text: 'See your PNR and ticket card in your profile.' },
];

function TransportArt({ type }) {
  const stroke = { fill: 'none', stroke: 'currentColor', strokeWidth: 3, strokeLinecap: 'round', strokeLinejoin: 'round' };

  if (type === 'train') {
    return (
      <svg viewBox="0 0 240 140" role="img" aria-label="Illustration of a train">
        <rect x="24" y="28" width="192" height="72" rx="22" fill="currentColor" opacity=".18" />
        <rect x="24" y="28" width="192" height="72" rx="22" {...stroke} />
        {[0, 1, 2].map((i) => (
          <rect key={i} x={44 + i * 46} y="42" width="34" height="26" rx="6" fill="currentColor" opacity=".55" />
        ))}
        <circle cx="198" cy="84" r="5" fill="currentColor" />
        <line x1="24" y1="84" x2="180" y2="84" {...stroke} />
        <line x1="8" y1="116" x2="232" y2="116" {...stroke} />
        {[20, 60, 100, 140, 180, 220].map((x) => (
          <line key={x} x1={x} y1="116" x2={x - 6} y2="128" {...stroke} strokeWidth="2" />
        ))}
      </svg>
    );
  }

  if (type === 'launch') {
    return (
      <svg viewBox="0 0 240 140" role="img" aria-label="Illustration of a launch (river boat)">
        <rect x="64" y="26" width="112" height="26" rx="6" fill="currentColor" opacity=".18" />
        <rect x="64" y="26" width="112" height="26" rx="6" {...stroke} />
        <rect x="44" y="56" width="152" height="28" rx="6" fill="currentColor" opacity=".18" />
        <rect x="44" y="56" width="152" height="28" rx="6" {...stroke} />
        {[0, 1, 2, 3, 4].map((i) => (
          <rect key={i} x={56 + i * 28} y="64" width="16" height="12" rx="3" fill="currentColor" opacity=".55" />
        ))}
        <path d="M20 90 H220 L198 118 H42 Z" fill="currentColor" opacity=".25" />
        <path d="M20 90 H220 L198 118 H42 Z" {...stroke} />
        <path d="M10 130 q15 -10 30 0 t30 0 t30 0 t30 0 t30 0 t30 0 t30 0" {...stroke} strokeWidth="2" />
      </svg>
    );
  }

  return (
    <svg viewBox="0 0 240 140" role="img" aria-label="Illustration of a bus">
      <rect x="20" y="28" width="200" height="74" rx="14" fill="currentColor" opacity=".18" />
      <rect x="20" y="28" width="200" height="74" rx="14" {...stroke} />
      {[0, 1, 2, 3].map((i) => (
        <rect key={i} x={36 + i * 40} y="42" width="30" height="26" rx="5" fill="currentColor" opacity=".55" />
      ))}
      <rect x="186" y="42" width="22" height="26" rx="5" fill="currentColor" opacity=".8" />
      <line x1="20" y1="84" x2="220" y2="84" {...stroke} />
      <circle cx="68" cy="104" r="13" fill="#0f172a" stroke="currentColor" strokeWidth="3" />
      <circle cx="172" cy="104" r="13" fill="#0f172a" stroke="currentColor" strokeWidth="3" />
      <line x1="8" y1="122" x2="232" y2="122" {...stroke} />
    </svg>
  );
}

function todayLocal() {
  const d = new Date();
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
  return d.toISOString().slice(0, 10);
}

export default function Home() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const urlType = searchParams.get('type');
  const serviceType = TYPES[urlType] ? urlType : 'bus';
  const current = TYPES[serviceType];

  const [fromCity, setFromCity] = useState('');
  const [toCity, setToCity] = useState('');
  const [journeyDate, setJourneyDate] = useState('');

  const [hint, setHint] = useState(null); // { field: 'from' | 'to', typed, suggestions }

  // Look the typed city up on the server. Returns null when it is fine (or the server cannot be
  // reached, in which case we do not block the search), otherwise a hint to show the user.
  const checkCity = async (field, typed) => {
    try {
      const res = await API.get('/cities', { params: { q: typed, limit: 3 } });
      const list = res.data || [];
      if (list.length > 0 && list[0].match === 'exact') return null;
      return { field, typed, suggestions: list.map((c) => c.name) };
    } catch {
      return null;
    }
  };

  const handleSearch = async (e) => {
    e.preventDefault();
    setHint(null);

    const problem = (await checkCity('from', fromCity.trim())) || (await checkCity('to', toCity.trim()));
    if (problem) {
      setHint(problem);
      return;
    }

    const params = new URLSearchParams({
      type: serviceType,
      from: fromCity.trim(),
      to: toCity.trim(),
      date: journeyDate,
    });
    navigate(`/search?${params.toString()}`);
  };

  const applySuggestion = (name) => {
    if (hint.field === 'from') setFromCity(name);
    else setToCity(name);
    setHint(null);
  };

  return (
    <div className="home" data-type={serviceType}>
      <section className="home-hero">
        <div className="home-hero-inner">
          <div>
            <span className="demo-badge">Academic demo</span>
            <h1>Plan your journey, from search to ticket</h1>
            <p className="home-sub">{current.tagline}</p>
          </div>

          <figure className="hero-art">
            <TransportArt type={serviceType} />
            <figcaption>Illustrative graphic · demo only</figcaption>
          </figure>
        </div>
      </section>

      <section className="search-card" aria-label="Search trips">
        <div className="type-tabs" role="tablist">
          {Object.entries(TYPES).map(([type, { label, Icon }]) => (
            <button
              key={type}
              type="button"
              role="tab"
              aria-selected={serviceType === type}
              className={serviceType === type ? 'type-tab active' : 'type-tab'}
              onClick={() => setSearchParams({ type })}
            >
              <Icon size={18} aria-hidden="true" /> {label}
            </button>
          ))}
        </div>

        <form onSubmit={handleSearch} className="search-grid">
          <CityInput id="from" label="From" icon={<MapPin size={14} aria-hidden="true" />}
            value={fromCity} onChange={setFromCity} placeholder="e.g. Dhaka" required />
          <CityInput id="to" label="To" icon={<MapPin size={14} aria-hidden="true" />}
            value={toCity} onChange={setToCity} placeholder="e.g. Sylhet" required />
          <div className="field">
            <label htmlFor="date"><Calendar size={14} aria-hidden="true" /> Date of journey</label>
            <input id="date" type="date" min={todayLocal()} value={journeyDate} onChange={(e) => setJourneyDate(e.target.value)} required />
          </div>
          <button type="submit" className="search-btn">
            <Search size={16} aria-hidden="true" /> Search {current.label}
          </button>
        </form>

        {hint && (
          <div className="city-hint" role="alert">
            {hint.suggestions.length > 0 ? (
              <>
                <span>We could not find &ldquo;{hint.typed}&rdquo; as {hint.field === 'from' ? 'a starting city' : 'a destination'}. Did you mean</span>
                {hint.suggestions.map((name) => (
                  <button key={name} type="button" onClick={() => applySuggestion(name)}>{name}</button>
                ))}
                <span>?</span>
              </>
            ) : (
              <span>
                We do not have any routes for &ldquo;{hint.typed}&rdquo; yet. Start typing and pick a city from the list.
              </span>
            )}
          </div>
        )}
      </section>

      <section className="steps" aria-labelledby="steps-title">
        <h2 id="steps-title">How this demo works</h2>
        <div className="steps-grid">
          {STEPS.map(({ Icon, title, text }, i) => (
            <div key={title} className="step">
              <span className="step-icon"><Icon size={20} aria-hidden="true" /></span>
              <h3>{i + 1}. {title}</h3>
              <p>{text}</p>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
