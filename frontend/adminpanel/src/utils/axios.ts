/**
 * axios setup to use mock service
 */

import axios, { AxiosRequestConfig } from 'axios';

import { dispatch } from 'store';
import { openSnackbar } from 'store/slices/snackbar';

declare global {
  interface Window {
    __CONFIG__?: { API_URL?: string; PROXY_URL?: string };
  }
}

const apiURL = window.__CONFIG__?.API_URL ?? import.meta.env.VITE_APP_API_URL ?? 'http://localhost:9091/';
const proxyURL = window.__CONFIG__?.PROXY_URL ?? import.meta.env.VITE_APP_PROXY_URL ?? 'http://localhost:8081/';

const axiosServices = axios.create({ baseURL: apiURL });
export const axiosProxyServices = axios.create({ baseURL: proxyURL });

// ==============================|| AXIOS - FOR MOCK SERVICES ||============================== //

axiosServices.interceptors.request.use(
  async (config) => {
    const accessToken = localStorage.getItem('serviceToken');
    if (accessToken) {
      config.headers['Authorization'] = `Bearer ${accessToken}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

axiosServices.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && !window.location.href.includes('/login')) {
      window.location.pathname = '/login';
    }
    if (error.response?.status === 403) {
      const message = error.response?.data?.error || 'You do not have permission to perform this action';
      dispatch(
        openSnackbar({
          open: true,
          message,
          variant: 'alert',
          alert: { variant: 'filled' },
          severity: 'error',
          close: true
        })
      );
    }
    return Promise.reject((error.response && error.response.data) || 'Wrong Services');
  }
);

export default axiosServices;

export async function fetcher(args: string | [string, AxiosRequestConfig]) {
  const [url, config] = Array.isArray(args) ? args : [args];

  const res = await axiosServices.get(url, { ...config });

  return res.data;
}
