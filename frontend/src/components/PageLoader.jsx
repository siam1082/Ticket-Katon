import { BusScene, TrainScene, LaunchScene } from './LoaderScenes';
import { useTransportType } from '../hooks/useTransportType';
import './PageLoader.css';

const SCENES = { bus: BusScene, train: TrainScene, launch: LaunchScene };

// Loading indicator used inside the app. It shows the vehicle the passenger is booking:
// a bus, a train or a launch. Pass type to force one; otherwise it follows ?type= in the
// address, or the last vehicle searched for.
// fullscreen: fills the page area (route changes). Otherwise it fits inside a card.
export default function PageLoader({ label = 'Loading…', fullscreen = false, type }) {
  const kind = useTransportType(type);
  const Scene = SCENES[kind];

  return (
    <div className={fullscreen ? 'pl pl--full' : 'pl'} role="status" aria-live="polite">
      <svg viewBox="0 0 300 110" role="img" aria-label={`A ${kind} on its way`}>
        <Scene />
      </svg>
      <p className="pl-label">{label}</p>
      <div className="pl-bar" aria-hidden="true"><i /></div>
    </div>
  );
}
