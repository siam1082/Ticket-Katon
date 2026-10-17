import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Bus, Train, Ship, Search, Calendar, MapPin } from 'lucide-react';

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
        <div style={styles.container}>
            <div style={styles.heroSection}>
                <h1 style={styles.heroTitle}>Book Tickets Easily for Any Journey</h1>
                <p style={styles.heroSubtitle}>Find bus, train, and launch schedules in one simple platform</p>
            </div>

            <div style={styles.searchCard}>
                {/* Service Type Switcher */}
                <div style={styles.tabBar}>
                    <button
                        type="button"
                        style={{
                            ...styles.tabBtn,
                            borderBottom: serviceType === 'bus' ? '3px solid #0066cc' : 'none',
                            color: serviceType === 'bus' ? '#0066cc' : '#4b5563',
                            fontWeight: serviceType === 'bus' ? '600' : '500'
                        }}
                        onClick={() => handleTypeChange('bus')}
                    >
                        <Bus size={18} /> Bus
                    </button>
                    <button
                        type="button"
                        style={{
                            ...styles.tabBtn,
                            borderBottom: serviceType === 'train' ? '3px solid #0066cc' : 'none',
                            color: serviceType === 'train' ? '#0066cc' : '#4b5563',
                            fontWeight: serviceType === 'train' ? '600' : '500'
                        }}
                        onClick={() => handleTypeChange('train')}
                    >
                        <Train size={18} /> Train
                    </button>
                    <button
                        type="button"
                        style={{
                            ...styles.tabBtn,
                            borderBottom: serviceType === 'launch' ? '3px solid #0066cc' : 'none',
                            color: serviceType === 'launch' ? '#0066cc' : '#4b5563',
                            fontWeight: serviceType === 'launch' ? '600' : '500'
                        }}
                        onClick={() => handleTypeChange('launch')}
                    >
                        <Ship size={18} /> Launch
                    </button>
                </div>

                {/* Search Input Controls */}
                <form onSubmit={handleSearch} style={styles.searchGrid}>
                    <div style={styles.inputGroup}>
                        <label style={styles.label}><MapPin size={14} /> From</label>
                        <input
                            type="text"
                            value={fromCity}
                            onChange={(e) => setFromCity(e.target.value)}
                            placeholder="e.g. Dhaka"
                            style={styles.input}
                            required
                        />
                    </div>

                    <div style={styles.inputGroup}>
                        <label style={styles.label}><MapPin size={14} /> To</label>
                        <input
                            type="text"
                            value={toCity}
                            onChange={(e) => setToCity(e.target.value)}
                            placeholder="e.g. Sylhet"
                            style={styles.input}
                            required
                        />
                    </div>

                    <div style={styles.inputGroup}>
                        <label style={styles.label}><Calendar size={14} /> Date of Journey</label>
                        <input
                            type="date"
                            value={journeyDate}
                            onChange={(e) => setJourneyDate(e.target.value)}
                            style={styles.input}
                            required
                        />
                    </div>

                    <div style={styles.buttonWrapper}>
                        <button type="submit" style={styles.searchSubmitBtn}>
                            <Search size={16} /> Search {serviceType.charAt(0).toUpperCase() + serviceType.slice(1)}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}

const styles = {
    container: { maxWidth: '1000px', margin: '0 auto', padding: '40px 16px' },
    heroSection: { textAlign: 'center', marginBottom: '32px' },
    heroTitle: { fontSize: '28px', color: '#111827', marginBottom: '8px' },
    heroSubtitle: { fontSize: '15px', color: '#6b7280' },
    searchCard: { backgroundColor: '#ffffff', borderRadius: '8px', boxShadow: '0 4px 14px rgba(0,0,0,0.06)', padding: '24px' },
    tabBar: { display: 'flex', gap: '24px', borderBottom: '1px solid #e5e7eb', marginBottom: '20px', paddingBottom: '8px' },
    tabBtn: { display: 'flex', alignItems: 'center', gap: '8px', background: 'none', border: 'none', cursor: 'pointer', fontSize: '15px', padding: '6px 4px' },
    searchGrid: { display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px', alignItems: 'end' },
    inputGroup: { display: 'flex', flexDirection: 'column', gap: '6px' },
    label: { fontSize: '13px', fontWeight: '500', color: '#374151', display: 'flex', alignItems: 'center', gap: '4px' },
    input: { padding: '10px 12px', border: '1px solid #d1d5db', borderRadius: '6px', fontSize: '14px', outline: 'none' },
    buttonWrapper: { display: 'flex', flexDirection: 'column', justifyContent: 'flex-end' },
    searchSubmitBtn: { display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px', padding: '11px', backgroundColor: '#0066cc', color: '#ffffff', border: 'none', borderRadius: '6px', cursor: 'pointer', fontWeight: '600' }
};