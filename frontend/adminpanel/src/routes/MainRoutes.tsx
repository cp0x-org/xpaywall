import { lazy } from 'react';

// project imports
import MainLayout from 'layout/MainLayout';
import Loadable from 'ui-component/Loadable';
import AuthGuard from 'utils/route-guard/AuthGuard';

// dashboard routing
const DashboardPage = Loadable(lazy(() => import('views/dashboard')));
const ProjectsPage = Loadable(lazy(() => import('views/projects')));
const ProjectFormPage = Loadable(lazy(() => import('views/projects/ProjectForm')));
const PaymentChannelsPage = Loadable(lazy(() => import('views/payment-channels')));
const RoutesPage = Loadable(lazy(() => import('views/routes-page')));
const RouteFormPage = Loadable(lazy(() => import('views/routes-page/RouteFormPage')));
const StatsPage = Loadable(lazy(() => import('views/stats')));
const RequestsPage = Loadable(lazy(() => import('views/requests')));
const EntityFormPage = Loadable(lazy(() => import('views/entity-pages/EntityFormPage')));

// ==============================|| MAIN ROUTING ||============================== //

const MainRoutes = {
  path: '/',
  element: (
    <AuthGuard>
      <MainLayout />
    </AuthGuard>
  ),
  children: [
    {
      path: '/dashboard',
      element: <DashboardPage />
    },
    {
      path: '/projects',
      element: <ProjectsPage />
    },
    {
      path: '/projects/create',
      element: <ProjectFormPage />
    },
    {
      path: '/projects/edit',
      element: <ProjectFormPage />
    },
    {
      path: '/projects/view',
      element: <ProjectFormPage />
    },
    {
      path: '/payment-channels',
      element: <PaymentChannelsPage />
    },
    {
      path: '/payment-channels/create',
      element: <EntityFormPage />
    },
    {
      path: '/payment-channels/edit',
      element: <EntityFormPage />
    },
    {
      path: '/payment-channels/view',
      element: <EntityFormPage />
    },
    {
      path: '/routes',
      element: <RoutesPage />
    },
    {
      path: '/routes/create',
      element: <RouteFormPage />
    },
    {
      path: '/routes/edit',
      element: <RouteFormPage />
    },
    {
      path: '/routes/view',
      element: <RouteFormPage />
    },
    {
      path: '/requests',
      element: <RequestsPage />
    },
    {
      path: '/stats',
      element: <StatsPage />
    },
    {
      path: '/stats/create',
      element: <EntityFormPage />
    },
    {
      path: '/stats/edit',
      element: <EntityFormPage />
    },
    {
      path: '/stats/view',
      element: <EntityFormPage />
    }
  ]
};

export default MainRoutes;
