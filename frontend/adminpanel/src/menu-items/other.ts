// assets
import { IconHelp, IconSitemap } from '@tabler/icons-react';
import { NavItemType } from 'types';

// constant
const icons = {
  IconHelp,
  IconSitemap
};

// ==============================|| SAMPLE PAGE & DOCUMENTATION MENU ITEMS ||============================== //

const other: NavItemType = {
  id: 'sample-docs-roadmap',
  icon: IconHelp,
  type: 'group',
  children: [
    {
      id: 'documentation',
      title: 'documentation',
      type: 'item',
      url: 'https://github.com/cp0x-org/xpaywall',
      icon: icons.IconHelp,
      external: true,
      target: true
    }
  ]
};

export default other;
