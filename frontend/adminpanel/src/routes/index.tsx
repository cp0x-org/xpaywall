import { Navigate } from 'react-router-dom';
import { createBrowserRouter } from 'react-router-dom';

// routes
import AuthenticationRoutes from './AuthenticationRoutes';
import LoginRoutes from './LoginRoutes';
import MainRoutes from './MainRoutes';

// ==============================|| ROUTING RENDER ||============================== //

const router = createBrowserRouter(
    [{ path: '/', element: <Navigate to="/dashboard" replace /> }, AuthenticationRoutes, LoginRoutes, MainRoutes],
    {
      basename: import.meta.env.VITE_APP_BASE_NAME
    }
);

export default router;
