import logo from '../assets/logo.svg';

/**
 * if you want to use image instead of <svg> uncomment following.
 *
 * import logoDark from 'assets/images/logo-dark.svg';
 * import logo from 'assets/images/logo.svg';
 *
 */

// ==============================|| LOGO SVG ||============================== //

export default function Logo({ dark = false }: { dark?: boolean }) {
  return (
    <div
      style={{
        display: 'inline-flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '3px'
      }}
    >
      <img
        src={logo}
        alt="Logo"
        width="63"
        height="26"
        style={{
          display: 'block',
          opacity: dark ? 0.92 : 1
        }}
      />
      <span
        style={{
          display: 'inline-flex',
          alignItems: 'baseline',
          justifyContent: 'center',
          gap: '1px',
          marginTop: '-1px',
          fontFamily: 'Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
        }}
      >
        <span
          style={{
            color: '#EEEEEE',
            fontSize: '1rem',
            fontWeight: 500,
            letterSpacing: 0,
            lineHeight: 1
          }}
        >
          x
        </span>
        <span
          style={{
            color: '#26cccb',
            fontSize: '1.1rem',
            fontWeight: 700,
            letterSpacing: 0,
            lineHeight: 1
          }}
        >
          Paywall
        </span>
      </span>
    </div>
  );
}
