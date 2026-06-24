import { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';

// material-ui
import Box from '@mui/material/Box';

// project imports
import useAuth from 'hooks/useAuth';
import { useDispatch } from 'store';
import { openSnackbar } from 'store/slices/snackbar';
import { DASHBOARD_PATH } from 'config';

// Google Identity Services is loaded on demand; the global is untyped.
declare global {
  interface Window {
    google?: any;
  }
}

const GIS_SRC = 'https://accounts.google.com/gsi/client';
const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID as string | undefined;

// loadGisScript injects the GIS script once and resolves when it is ready.
function loadGisScript(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (window.google?.accounts?.id) {
      resolve();
      return;
    }
    const existing = document.querySelector<HTMLScriptElement>(`script[src="${GIS_SRC}"]`);
    if (existing) {
      existing.addEventListener('load', () => resolve());
      existing.addEventListener('error', () => reject(new Error('Failed to load Google script')));
      return;
    }
    const script = document.createElement('script');
    script.src = GIS_SRC;
    script.async = true;
    script.defer = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error('Failed to load Google script'));
    document.body.appendChild(script);
  });
}

// ==============================|| GOOGLE SIGN-IN BUTTON ||============================== //

export default function GoogleLoginButton() {
  const { googleLogin } = useAuth();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const buttonRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Without a configured client id there is nothing to render.
    if (!clientId) return;

    let cancelled = false;

    loadGisScript()
      .then(() => {
        if (cancelled || !buttonRef.current) return;

        window.google.accounts.id.initialize({
          client_id: clientId,
          callback: async (response: { credential?: string }) => {
            if (!response.credential) return;
            try {
              await googleLogin?.(response.credential);
              navigate(DASHBOARD_PATH, { replace: true });
            } catch (err: any) {
              const message = err?.error || err?.message || 'Google sign-in failed';
              dispatch(
                openSnackbar({
                  open: true,
                  message,
                  variant: 'alert',
                  alert: { variant: 'filled' },
                  severity: 'error',
                  anchorOrigin: { vertical: 'top', horizontal: 'center' },
                  close: true
                })
              );
            }
          }
        });

        window.google.accounts.id.renderButton(buttonRef.current, {
          theme: 'outline',
          size: 'large',
          width: 320,
          text: 'continue_with'
        });
      })
      .catch((err) => console.error(err));

    return () => {
      cancelled = true;
    };
  }, [googleLogin, navigate, dispatch]);

  if (!clientId) return null;

  return <Box ref={buttonRef} sx={{ display: 'flex', justifyContent: 'center' }} />;
}
