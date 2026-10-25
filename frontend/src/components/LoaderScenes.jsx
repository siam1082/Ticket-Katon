// The three loader animations. Each one draws a vehicle travelling through its own surroundings.
// The moving parts are animated in PageLoader.css.

const SKYLINE = (id) => (
  <pattern id={id} width="60" height="40" patternUnits="userSpaceOnUse">
    <rect x="2" y="14" width="14" height="26" fill="currentColor" opacity=".12" />
    <rect x="20" y="4" width="12" height="36" fill="currentColor" opacity=".18" />
    <rect x="36" y="20" width="10" height="20" fill="currentColor" opacity=".12" />
    <rect x="48" y="10" width="10" height="30" fill="currentColor" opacity=".15" />
  </pattern>
);

function Wheel({ cx, cy, r = 8 }) {
  return (
    <g className="pl-wheel">
      <circle cx={cx} cy={cy} r={r} fill="#0f172a" stroke="#99f6e4" strokeWidth="2" />
      <line x1={cx} y1={cy - r + 3} x2={cx} y2={cy + r - 3} stroke="#99f6e4" strokeWidth="1.5" />
      <line x1={cx - r + 3} y1={cy} x2={cx + r - 3} y2={cy} stroke="#99f6e4" strokeWidth="1.5" />
    </g>
  );
}

export function BusScene() {
  return (
    <>
      <defs>{SKYLINE('pl-skyline')}</defs>
      <rect className="pl-sky" x="0" y="46" width="360" height="40" fill="url(#pl-skyline)" />
      <g className="pl-bus">
        <rect x="104" y="46" width="108" height="38" rx="9" fill="#0f766e" />
        {[112, 132, 152, 172].map((x) => (
          <rect key={x} x={x} y="53" width="16" height="12" rx="3" fill="#ccfbf1" />
        ))}
        <rect x="192" y="53" width="14" height="22" rx="3" fill="#99f6e4" />
        <line x1="104" y1="72" x2="212" y2="72" stroke="#115e59" strokeWidth="2" />
        <Wheel cx={132} cy={86} />
        <Wheel cx={186} cy={86} />
      </g>
      <line className="pl-road" x1="-20" y1="98" x2="320" y2="98" stroke="#0f766e" strokeWidth="3" strokeLinecap="round" strokeDasharray="14 12" opacity=".55" />
    </>
  );
}

export function TrainScene() {
  return (
    <>
      <defs>{SKYLINE('pl-skyline-t')}</defs>
      <rect className="pl-sky" x="0" y="40" width="360" height="40" fill="url(#pl-skyline-t)" />
      <g className="pl-bus">
        {/* two carriages */}
        {[36, 108].map((x) => (
          <g key={x}>
            <rect x={x} y="50" width="68" height="34" rx="6" fill="#0f766e" />
            {[8, 22, 36, 50].map((dx) => (
              <rect key={dx} x={x + dx} y="57" width="10" height="10" rx="2" fill="#ccfbf1" />
            ))}
            <line x1={x} y1="74" x2={x + 68} y2="74" stroke="#115e59" strokeWidth="2" />
          </g>
        ))}
        {/* engine */}
        <path d="M180 50h46c16 0 32 12 36 34h-82z" fill="#115e59" />
        <rect x="190" y="57" width="24" height="12" rx="2" fill="#ccfbf1" />
        <circle cx="257" cy="76" r="2.5" fill="#fde68a" />
        {/* couplers */}
        <rect x="104" y="72" width="4" height="4" fill="#115e59" />
        <rect x="176" y="72" width="4" height="4" fill="#115e59" />
        {[52, 88, 124, 160, 198, 240].map((cx) => (
          <Wheel key={cx} cx={cx} cy={88} r={5} />
        ))}
      </g>
      <line x1="-20" y1="94.5" x2="320" y2="94.5" stroke="#0f766e" strokeWidth="2" opacity=".6" />
      <line className="pl-sleepers" x1="-20" y1="100" x2="320" y2="100" stroke="#0f766e" strokeWidth="4" strokeDasharray="5 9" opacity=".4" />
    </>
  );
}

export function LaunchScene() {
  return (
    <>
      <defs>
        <pattern id="pl-trees" width="60" height="30" patternUnits="userSpaceOnUse">
          <circle cx="12" cy="16" r="9" fill="currentColor" opacity=".14" />
          <circle cx="34" cy="12" r="12" fill="currentColor" opacity=".18" />
          <circle cx="52" cy="18" r="8" fill="currentColor" opacity=".12" />
        </pattern>
        <pattern id="pl-wave-p" width="30" height="8" patternUnits="userSpaceOnUse">
          <path d="M0 4q7.5-6 15 0t15 0" fill="none" stroke="currentColor" strokeWidth="2" opacity=".5" />
        </pattern>
      </defs>
      <rect className="pl-sky" x="0" y="46" width="360" height="30" fill="url(#pl-trees)" />
      <rect x="0" y="78" width="300" height="32" fill="#0f766e" opacity=".1" />
      <g className="pl-boat">
        <path d="M92 76h124l-16 15h-92z" fill="#0f766e" />
        <rect x="104" y="62" width="96" height="14" rx="3" fill="#99f6e4" />
        {[112, 128, 144, 160, 176].map((x) => (
          <rect key={x} x={x} y="66" width="9" height="6" rx="1.5" fill="#0f766e" opacity=".75" />
        ))}
        <rect x="120" y="50" width="64" height="12" rx="3" fill="#ccfbf1" />
        <rect x="140" y="40" width="26" height="10" rx="2" fill="#99f6e4" />
        <rect x="150" y="29" width="8" height="11" fill="#115e59" />
        <g className="pl-smoke" fill="#94a3b8">
          <circle cx="154" cy="26" r="3" />
          <circle cx="154" cy="26" r="3" />
          <circle cx="154" cy="26" r="3" />
        </g>
      </g>
      <rect className="pl-wave" x="0" y="87" width="360" height="8" fill="url(#pl-wave-p)" />
      <rect className="pl-wave pl-wave--slow" x="0" y="96" width="360" height="8" fill="url(#pl-wave-p)" />
    </>
  );
}
