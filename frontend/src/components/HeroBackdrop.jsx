// A faint city, train, bus and river launch drawn along the bottom of the hero.
// It is only decoration, so it is hidden from screen readers.
export default function HeroBackdrop() {
  return (
    <svg className="hero-backdrop" viewBox="0 0 1440 300" preserveAspectRatio="xMidYMax slice" aria-hidden="true" focusable="false">
      {/* skyline */}
      <g fill="#fff" opacity=".07">
        <rect x="40" y="170" width="60" height="130" />
        <rect x="110" y="120" width="46" height="180" />
        <rect x="170" y="190" width="70" height="110" />
        <rect x="480" y="150" width="54" height="150" />
        <rect x="545" y="196" width="80" height="104" />
        <rect x="860" y="130" width="50" height="170" />
        <rect x="920" y="180" width="72" height="120" />
        <rect x="1010" y="112" width="44" height="188" />
        <rect x="1180" y="168" width="66" height="132" />
        <rect x="1256" y="140" width="48" height="160" />
        <rect x="1316" y="196" width="84" height="104" />
      </g>

      {/* train */}
      <g>
        <rect x="260" y="206" width="330" height="62" rx="26" fill="#fff" opacity=".12" />
        <g fill="#000" opacity=".1">
          <rect x="292" y="220" width="40" height="24" rx="5" />
          <rect x="344" y="220" width="40" height="24" rx="5" />
          <rect x="396" y="220" width="40" height="24" rx="5" />
          <rect x="448" y="220" width="40" height="24" rx="5" />
          <rect x="500" y="220" width="40" height="24" rx="5" />
        </g>
      </g>

      {/* bus */}
      <g>
        <rect x="660" y="214" width="170" height="54" rx="12" fill="#fff" opacity=".12" />
        <g fill="#000" opacity=".1">
          <rect x="674" y="224" width="26" height="20" rx="4" />
          <rect x="708" y="224" width="26" height="20" rx="4" />
          <rect x="742" y="224" width="26" height="20" rx="4" />
          <rect x="776" y="224" width="40" height="30" rx="4" />
        </g>
        <circle cx="704" cy="270" r="13" fill="#fff" opacity=".16" />
        <circle cx="786" cy="270" r="13" fill="#fff" opacity=".16" />
      </g>

      {/* launch on the river */}
      <g>
        <path d="M1030 244h250l-30 38h-190z" fill="#fff" opacity=".12" />
        <rect x="1062" y="214" width="190" height="30" rx="5" fill="#fff" opacity=".1" />
        <rect x="1100" y="190" width="116" height="24" rx="5" fill="#fff" opacity=".08" />
      </g>

      {/* ground */}
      <rect x="0" y="284" width="1440" height="16" fill="#fff" opacity=".1" />
    </svg>
  );
}
