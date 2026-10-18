import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Bus, Train, Ship, Search, Calendar, MapPin, ArrowRight, ShieldCheck, Ticket, Users } from 'lucide-react';
import './Home.css';

export default function Home() {
    const [searchParams, setSearchParams] = useSearchParams();
    const activeTypeFromUrl = searchParams.get('type') || 'bus';

    const [serviceType, setServiceType] = useState(activeTypeFromUrl);
    const [fromCity, setFromCity] = useState('');
    const [toCity, setToCity] = useState('');
    const [journeyDate, setJourneyDate] = useState('');

    useEffect(() => {
        const type = searchParams.get('type');
        if (type && ['bus', 'train', 'launch'].includes(type)) {
            setServiceType(type);
        }
    }, [searchParams]);

    const handleTypeChange = (type) => {
        setServiceType(type);
        setSearchParams({ type });
    };

    const handleSearch = (e) => {
        e.preventDefault();
        alert(`Searching ${serviceType.toUpperCase()} tickets from ${fromCity} to ${toCity} on ${journeyDate}`);
    };

    return (
        <div className="gov-home-shell">
            <section className="gov-hero">
                <div className="gov-hero-content">
                    <span className="gov-kicker">Official Travel Services</span>
                    <h1>Book Tickets Easily for Any Journey</h1>
                    <p>
                        Find bus, train, and launch schedules in one secure government-backed booking platform.
                    </p>
                    <div className="gov-hero-actions">
                        <button type="button" className="gov-primary-btn">Book a Ticket <ArrowRight size={16} /></button>
                        <button type="button" className="gov-secondary-btn">View Schedules</button>
                    </div>
                </div>

                <div className="gov-hero-stats">
                    <div className="stat-item">
                        <Ticket size={18} />
                        <div>
                            <strong>12K+</strong>
                            <span>Tickets issued</span>
                        </div>
                    </div>
                    <div className="stat-item">
                        <Users size={18} />
                        <div>
                            <strong>3.4K</strong>
                            <span>Daily passengers</span>
                        </div>
                    </div>
                    <div className="stat-item">
                        <ShieldCheck size={18} />
                        <div>
                            <strong>100%</strong>
                            <span>Verified service</span>
                        </div>
                    </div>
                </div>
            </section>

            <section className="gov-search-card">
                <div className="gov-tab-bar">
                    <button
                        type="button"
                        className={serviceType === 'bus' ? 'gov-tab-btn active' : 'gov-tab-btn'}
                        onClick={() => handleTypeChange('bus')}
                    >
                        <Bus size={18} /> Bus
                    </button>
                    <button
                        type="button"
                        className={serviceType === 'train' ? 'gov-tab-btn active' : 'gov-tab-btn'}
                        onClick={() => handleTypeChange('train')}
                    >
                        <Train size={18} /> Train
                    </button>
                    <button
                        type="button"
                        className={serviceType === 'launch' ? 'gov-tab-btn active' : 'gov-tab-btn'}
                        onClick={() => handleTypeChange('launch')}
                    >
                        <Ship size={18} /> Launch
                    </button>
                </div>

                <form onSubmit={handleSearch} className="gov-search-grid">
                    <div className="gov-input-group">
                        <label><MapPin size={14} /> From</label>
                        <input
                            type="text"
                            value={fromCity}
                            onChange={(e) => setFromCity(e.target.value)}
                            placeholder="e.g. Dhaka"
                            required
                        />
                    </div>

                    <div className="gov-input-group">
                        <label><MapPin size={14} /> To</label>
                        <input
                            type="text"
                            value={toCity}
                            onChange={(e) => setToCity(e.target.value)}
                            placeholder="e.g. Sylhet"
                            required
                        />
                    </div>

                    <div className="gov-input-group">
                        <label><Calendar size={14} /> Date of Journey</label>
                        <input
                            type="date"
                            value={journeyDate}
                            onChange={(e) => setJourneyDate(e.target.value)}
                            required
                        />
                    </div>

                    <div className="gov-button-wrap">
                        <button type="submit" className="gov-search-btn">
                            <Search size={16} /> Search {serviceType.charAt(0).toUpperCase() + serviceType.slice(1)}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    );
}
