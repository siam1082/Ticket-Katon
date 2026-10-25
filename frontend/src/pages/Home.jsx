import { useEffect, useState } from 'react';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';
import {
  Bus, Train, Ship, Search, Calendar, MapPin, Armchair, CreditCard, Ticket,
  ArrowLeftRight, ArrowRight, Timer, Undo2,
} from 'lucide-react';
import CityInput from '../components/CityInput';
import HeroBackdrop from '../components/HeroBackdrop';
import TeamSection from '../components/TeamSection';
import API from '../services/api';
import { rememberTransportType } from '../hooks/useTransportType';
import './Home.css';

const TYPES = {
  bus: { label: 'Bus', Icon: Bus },
  train: { label: 'Train', Icon: Train },
  launch: { label: 'Launch', Icon: Ship },
};

const STEPS = [
  { Icon: Search, title: 'Search', text: 'Choose your route and travel date to see the available trips.' },
  { Icon: Armchair, title: 'Select your seat', text: 'Pick a free seat on the live seat map. It is held for you for 5 minutes.' },
  { Icon: CreditCard, title: 'Demo payment', text: 'Complete a simulated payment. No real money is used.' },
  { Icon: Ticket, title: 'Get your ticket', text: 'Your PNR and e-ticket appear in your profile straight away.' },
];

const FEATURES = [
  { Icon: Armchair, title: 'Live seat maps', text: 'See which seats are free, held or booked, for bus, train and launch.' },
  { Icon: Timer, title: '5-minute seat hold', text: 'Your seat is kept for you while you pay, with a countdown on screen.' },
  { Icon: Ticket, title: 'Instant e-ticket', text: 'Your ticket is ready as soon as the demo payment succeeds.' },
  { Icon: Undo2, title: 'Easy cancellation', text: 'See your refund before you confirm. Cancelled seats go back on sale at once.' },
];

const POPULAR_ROUTES = [
  { type: 'bus', from: 'Dhaka', to: 'Sylhet' },
  { type: 'train', from: 'Dhaka', to: 'Chattogram' },
  { type: 'launch', from: 'Dhaka', to: 'Barishal' },
  { type: 'bus', from: 'Sylhet', to: 'Dhaka' },
  { type: 'train', from: 'Chattogram', to: 'Dhaka' },
  { type: 'launch', from: 'Barishal', to: 'Dhaka' },
];

const QUICK_DATES = [
  { label: 'Today', offset: 0 },
  { label: 'Tomorrow', offset: 1 },
  { label: 'In 2 days', offset: 2 },
];

function dateFromToday(offset) {
  const d = new Date();
  d.setDate(d.getDate() + offset);
  d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
  return d.toISOString().slice(0, 10);
}

export default function Home() {
  const navigate = useNavigate();
  const { hash } = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();
  const urlType = searchParams.get('type');
  const serviceType = TYPES[urlType] ? urlType : 'bus';
  const current = TYPES[serviceType];

  const [fromCity, setFromCity] = useState('');
  const [toCity, setToCity] = useState('');
  const [journeyDate, setJourneyDate] = useState('');
  const [hint, setHint] = useState(null); // { field: 'from' | 'to', typed, suggestions }
  const [searching, setSearching] = useState(false);

  // Links like /#team scroll to that section; a plain "/" goes back to the top.
  useEffect(() => {
    if (hash) document.getElementById(hash.slice(1))?.scrollIntoView({ behavior: 'smooth' });
    else window.scrollTo({ top: 0 });
  }, [hash]);

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

  const goToResults = (type, from, to, date) => {
    rememberTransportType(type);
    const params = new URLSearchParams({ type, from, to, date });
    navigate(`/search?${params.toString()}`);
  };

  const handleSearch = async (e) => {
    e.preventDefault();
    setHint(null);
    setSearching(true);
    try {
      const problem = (await checkCity('from', fromCity.trim())) || (await checkCity('to', toCity.trim()));
      if (problem) {
        setHint(problem);
        return;
      }
      goToResults(serviceType, fromCity.trim(), toCity.trim(), journeyDate);
    } finally {
      setSearching(false);
    }
  };

  const applySuggestion = (name) => {
    if (hint.field === 'from') setFromCity(name);
    else setToCity(name);
    setHint(null);
  };

  const swapCities = () => {
    setFromCity(toCity);
    setToCity(fromCity);
    setHint(null);
  };

  return (
    <div className="home">
      <section className="home-hero">
        <HeroBackdrop />

        <div className="hero-inner">
          <span className="demo-badge">Academic demo</span>
          <h1>Book Bus, Train &amp; Launch Tickets</h1>
          <p className="home-sub">Search your route, choose your seat and get your e-ticket in minutes.</p>

          <div className="search-wrap">
            <div className="type-tabs" role="tablist" aria-label="Transport type">
              {Object.entries(TYPES).map(([type, { label, Icon }]) => (
                <button
                  key={type}
                  type="button"
                  role="tab"
                  aria-selected={serviceType === type}
                  className={serviceType === type ? 'type-tab active' : 'type-tab'}
                  onClick={() => {
                    rememberTransportType(type);
                    setSearchParams({ type });
                  }}
                >
                  <Icon size={17} aria-hidden="true" /> {label}
                </button>
              ))}
            </div>

            <div className="search-card">
              <form onSubmit={handleSearch} className="search-grid">
                <CityInput id="from" label="From" icon={<MapPin size={14} aria-hidden="true" />}
                  value={fromCity} onChange={setFromCity} placeholder="Select departure city" required />

                <button type="button" className="swap-btn" onClick={swapCities} aria-label="Swap From and To">
                  <ArrowLeftRight size={16} aria-hidden="true" />
                </button>

                <CityInput id="to" label="To" icon={<MapPin size={14} aria-hidden="true" />}
                  value={toCity} onChange={setToCity} placeholder="Select arrival city" required />

                <div className="field">
                  <label htmlFor="date"><Calendar size={14} aria-hidden="true" /> Date of journey</label>
                  <input id="date" type="date" min={dateFromToday(0)} value={journeyDate}
                    onChange={(e) => setJourneyDate(e.target.value)} required />
                </div>

                <button type="submit" className="search-btn" disabled={searching}>
                  {searching ? <span className="spinner" aria-hidden="true" /> : <Search size={16} aria-hidden="true" />}
                  {searching ? 'Searching...' : `Search ${current.label}`}
                </button>
              </form>

              <div className="quick-dates">
                <span>Quick pick</span>
                {QUICK_DATES.map(({ label, offset }) => {
                  const value = dateFromToday(offset);
                  return (
                    <button key={label} type="button" className={journeyDate === value ? 'chip on' : 'chip'}
                      onClick={() => setJourneyDate(value)}>
                      {label}
                    </button>
                  );
                })}
              </div>

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
            </div>
          </div>
        </div>
      </section>

      <section className="section" id="how" aria-labelledby="how-title">
        <div className="section-head">
          <h2 id="how-title">How it works</h2>
          <p>Four simple steps from search to ticket.</p>
        </div>
        <div className="steps-grid">
          {STEPS.map(({ Icon, title, text }, i) => (
            <div key={title} className="step">
              <span className="step-no">{String(i + 1).padStart(2, '0')}</span>
              <span className="step-icon"><Icon size={22} aria-hidden="true" /></span>
              <h3>{title}</h3>
              <p>{text}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="section" id="routes" aria-labelledby="routes-title">
        <div className="section-head">
          <h2 id="routes-title">Popular routes</h2>
          <p>Choose a route to see trips for tomorrow.</p>
        </div>
        <div className="route-grid">
          {POPULAR_ROUTES.map((r) => {
            const { Icon, label } = TYPES[r.type];
            return (
              <button
                key={`${r.type}-${r.from}-${r.to}`}
                type="button"
                className="route-card"
                onClick={() => goToResults(r.type, r.from, r.to, dateFromToday(1))}
              >
                <span className="route-icon"><Icon size={20} aria-hidden="true" /></span>
                <span className="route-text">
                  <strong>{r.from} <ArrowRight size={14} aria-hidden="true" /> {r.to}</strong>
                  <small>{label} · daily</small>
                </span>
                <span className="route-go">View trips <ArrowRight size={14} aria-hidden="true" /></span>
              </button>
            );
          })}
        </div>
      </section>

      <section className="section" aria-labelledby="why-title">
        <div className="section-head">
          <h2 id="why-title">Why TicketKaton</h2>
          <p>Everything you need to book with confidence.</p>
        </div>
        <div className="feature-grid">
          {FEATURES.map(({ Icon, title, text }) => (
            <div key={title} className="feature">
              <span className="feature-icon"><Icon size={22} aria-hidden="true" /></span>
              <h3>{title}</h3>
              <p>{text}</p>
            </div>
          ))}
        </div>
      </section>

      <TeamSection />
    </div>
  );
}
