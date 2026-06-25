import React, { createContext, useEffect, useReducer } from 'react';

// third party
import { jwtDecode } from 'jwt-decode';

// reducer - state management
import { LOGIN, LOGOUT } from 'store/actions';
import accountReducer from 'store/accountReducer';

// project imports
import Loader from 'ui-component/Loader';
import axios from 'utils/axios';

// types
import { KeyedObject } from 'types';
import { InitialLoginContextProps, JWTContextType } from 'types/auth';

// constant
const initialState: InitialLoginContextProps = {
  isLoggedIn: false,
  isInitialized: false,
  user: null
};

function verifyToken(serviceToken: string): boolean {
  if (!serviceToken) {
    return false;
  }

  const decoded: KeyedObject = jwtDecode(serviceToken);

  if (!decoded.exp) {
    throw new Error("Token does not contain 'exp' property.");
  }

  return decoded.exp > Date.now() / 1000;
}

function setSession(serviceToken?: string | null): void {
  if (serviceToken) {
    localStorage.setItem('serviceToken', serviceToken);
    axios.defaults.headers.common.Authorization = `Bearer ${serviceToken}`;
  } else {
    localStorage.removeItem('serviceToken');
    delete axios.defaults.headers.common.Authorization;
  }
}

// ==============================|| JWT CONTEXT & PROVIDER ||============================== //

const JWTContext = createContext<JWTContextType | null>(null);

export function JWTProvider({ children }: { children: React.ReactElement }) {
  const [state, dispatch] = useReducer(accountReducer, initialState);

  useEffect(() => {
    const init = async () => {
      try {
        const serviceToken = window.localStorage.getItem('serviceToken');
        if (serviceToken && verifyToken(serviceToken)) {
          setSession(serviceToken);
          const response = await axios.get('/auth/me');
          const user = response.data;
          dispatch({
            type: LOGIN,
            payload: {
              isLoggedIn: true,
              user
            }
          });
        } else {
          dispatch({
            type: LOGOUT
          });
        }
      } catch (err) {
        console.error(err);
        dispatch({
          type: LOGOUT
        });
      }
    };

    init();
  }, []);

  // identifier may be a username or an email address; the backend resolves either.
  const login = async (identifier: string, password: string) => {
    const response = await axios.post('/auth/login', { username: identifier, password });
    const { token, user } = response.data;
    setSession(token);
    dispatch({
      type: LOGIN,
      payload: {
        isLoggedIn: true,
        user
      }
    });
  };

  // register creates a local account and signs the user in (backend returns a token).
  const register = async (username: string, email: string, password: string) => {
    const response = await axios.post('/auth/register', { username, email, password });
    const { token, user } = response.data;
    setSession(token);
    dispatch({
      type: LOGIN,
      payload: {
        isLoggedIn: true,
        user
      }
    });
  };

  // googleLogin verifies a Google ID token (from Google Identity Services) and signs in.
  const googleLogin = async (idToken: string) => {
    const response = await axios.post('/auth/google', { id_token: idToken });
    const { token, user } = response.data;
    setSession(token);
    dispatch({
      type: LOGIN,
      payload: {
        isLoggedIn: true,
        user
      }
    });
  };

  const logout = () => {
    setSession(null);
    dispatch({ type: LOGOUT });
  };

  // requestPasswordReset asks for a reset link. The backend always returns a
  // generic message; in dev (no SMTP) it also returns the link directly.
  const requestPasswordReset = async (email: string): Promise<{ message: string; resetUrl?: string }> => {
    const response = await axios.post('/auth/forgot-password', { email });
    return { message: response.data?.message ?? '', resetUrl: response.data?.reset_url };
  };

  // confirmResetPassword consumes a reset token, sets a new password and returns
  // the backend's confirmation message.
  const confirmResetPassword = async (token: string, password: string): Promise<string> => {
    const response = await axios.post('/auth/reset-password', { token, password });
    return response.data?.message ?? 'Password updated';
  };

  // resetPassword kept for context-type compatibility; delegates to requestPasswordReset.
  const resetPassword = async (email: string) => {
    await requestPasswordReset(email);
  };

  const updateProfile = () => {};

  if (state.isInitialized !== undefined && !state.isInitialized) {
    return <Loader />;
  }

  return (
    <JWTContext
      value={{ ...state, login, logout, register, googleLogin, requestPasswordReset, confirmResetPassword, resetPassword, updateProfile }}
    >
      {children}
    </JWTContext>
  );
}

export default JWTContext;
