// This is example of menu item without group for horizontal layout. There will be no children.

// assets
import {
  IconApps,
  IconUserCheck,
  IconBasket,
  IconFileInvoice,
  IconMessages,
  IconLayoutKanban,
  IconMail,
  IconCalendar,
  IconNfc,
  IconHeadphones,
  IconArticle,
  IconLifebuoy,
  IconDashboard,
  IconBrandChrome,
  IconRoute2,
  IconCapProjecting, IconFolderBolt, IconMoneybagEdit, IconMoneybagHeart, IconMoneybagPlus,
  IconCoins, IconServer,
  IconList
} from '@tabler/icons-react';

// types
import { NavItemType } from 'types';

// ==============================|| MENU ITEMS - SAMPLE PAGE ||============================== //

const icons = {
  IconBrandChrome,
  IconDashboard,
  IconBasket,
  IconRoute2,
  IconCapProjecting
};
const mainPage: NavItemType = {
  id: 'xpaywall',
  icon: icons.IconBrandChrome,
  type: 'group',
  children: [
    {
      id: 'dashboard',
      title: 'Dashboard',
      type: 'item',
      url: '/dashboard',
      icon: IconDashboard,
      breadcrumbs: false,
    },
    {
      id: 'routes-item',
      title: 'Routes',
      type: 'item',
      url: '/routes',
      icon: icons.IconRoute2,
    },
    {
      id: 'projects',
      title: 'Projects',
      type: 'collapse',
      icon: IconFolderBolt,
      children: [
        {
          id: 'project-list',
          title: 'Project List',
          type: 'item',
          url: '/projects',
          icon: IconCapProjecting,
        },
        {
          id: 'payment-methods',
          title: 'Payment Methods',
          type: 'item',
          url: '/project-payment-methods',
          icon: IconMoneybagHeart,
        },
      ]
    },
    {
      id: 'payment-channels-item',
      title: 'Payments',
      type: 'collapse',
      // url: '/payment-channels',
      icon: IconMoneybagPlus,
      children: [
        {
          id: 'payment-method-item',
          title: 'Payment Methods',
          type: 'item',
          url: '/payment-methods',
          icon: IconMoneybagPlus,
        },
        {
          id: 'payment-assets-item',
          title: 'Payment Assets',
          type: 'item',
          url: '/payment-assets',
          icon: IconCoins,
        },
        {
          id: 'facilitators',
          title: 'Facilitators (x402)',
          type: 'item',
          url: '/facilitators',
          icon: IconServer,
        },
      ],
    },
    {
      id: 'requests-item',
      title: 'Requests',
      type: 'item',
      url: '/requests',
      icon: IconList,
    },
    // {
    //   id: 'stats-item',
    //   title: 'Stats',
    //   type: 'item',
    //   url: '/stats'
    // }
  ]
};

export default mainPage;
