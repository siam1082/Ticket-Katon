import { useLocation } from 'react-router-dom';

const KEY = 'tk:type';
export const TRANSPORT_TYPES = ['bus', 'train', 'launch'];

const clean = (value) => String(value || '').toLowerCase();

// Remember the vehicle the passenger is booking, so a loader can show the same one.
export function rememberTransportType(type) {
  const t = clean(type);
  if (!TRANSPORT_TYPES.includes(t)) return;
  try {
    sessionStorage.setItem(KEY, t);
  } catch {
    // Storage can be blocked; the loader then falls back to the address or the bus.
  }
}

function readRemembered() {
  try {
    const t = sessionStorage.getItem(KEY);
    return TRANSPORT_TYPES.includes(t) ? t : null;
  } catch {
    return null;
  }
}

// Which vehicle to show: the one passed in, else ?type= in the address, else the last one
// the passenger searched for, else a bus.
export function useTransportType(explicit) {
  const { search } = useLocation();
  const fromUrl = new URLSearchParams(search).get('type');
  const found = [explicit, fromUrl].map(clean).find((t) => TRANSPORT_TYPES.includes(t));
  return found || readRemembered() || 'bus';
}
