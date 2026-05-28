import { lazy } from 'react';

// project imports
import MainLayout from 'layout/MainLayout';
import Loadable from 'ui-component/Loadable';
import AuthGuard from 'utils/route-guard/AuthGuard';

// dashboard routing
const DashboardPage = Loadable(lazy(() => import('views/dashboard')));
const ProjectsPage = Loadable(lazy(() => import('views/projects')));
const ProjectFormPage = Loadable(lazy(() => import('views/projects/ProjectForm')));
const PaymentChannelsPage = Loadable(lazy(() => import('views/payment-methods')));
const PaymentMethodForm = Loadable(lazy(() => import('views/payment-methods/PaymentMethodForm')));
const PaymentAssetsPage = Loadable(lazy(() => import('views/payment-assets')));
const PaymentAssetForm = Loadable(lazy(() => import('views/payment-assets/PaymentAssetForm')));
const FacilitatorsPage = Loadable(lazy(() => import('views/facilitators')));
const FacilitatorForm = Loadable(lazy(() => import('views/facilitators/FacilitatorForm')));
const RoutesPage = Loadable(lazy(() => import('views/routes-page')));
const RouteFormPage = Loadable(lazy(() => import('views/routes-page/RouteFormPage')));
const StatsPage = Loadable(lazy(() => import('views/stats')));
const ProjectPaymentMethodsPage = Loadable(lazy(() => import('views/project-payment-methods')));
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
      path: '/payment-methods',
      element: <PaymentChannelsPage />
    },
    {
      path: '/payment-methods/create',
      element: <PaymentMethodForm />
    },
    {
      path: '/payment-methods/edit',
      element: <PaymentMethodForm />
    },
    {
      path: '/payment-methods/view',
      element: <PaymentMethodForm />
    },
    {
      path: '/payment-assets',
      element: <PaymentAssetsPage />
    },
    {
      path: '/payment-assets/create',
      element: <PaymentAssetForm />
    },
    {
      path: '/payment-assets/edit',
      element: <PaymentAssetForm />
    },
    {
      path: '/payment-assets/view',
      element: <PaymentAssetForm />
    },
    {
      path: '/facilitators',
      element: <FacilitatorsPage />
    },
    {
      path: '/facilitators/create',
      element: <FacilitatorForm />
    },
    {
      path: '/facilitators/edit',
      element: <FacilitatorForm />
    },
    {
      path: '/facilitators/view',
      element: <FacilitatorForm />
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
      path: '/project-payment-methods',
      element: <ProjectPaymentMethodsPage />
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
