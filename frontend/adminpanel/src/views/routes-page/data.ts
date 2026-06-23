import { EntityRow } from 'views/entity-pages/types';

export const ROUTE_ROWS: EntityRow[] = [
  {
    id: 3001,
    name: 'US Primary Route',
    owner: 'Amelia Price',
    status: 'Active',
    startDate: '2026-01-08',
    dueDate: '2026-06-21',
    budget: 31000
  },
  {
    id: 3002,
    name: 'EU Fallback Route',
    owner: 'George Mills',
    status: 'Planned',
    startDate: '2026-05-25',
    dueDate: '2026-09-15',
    budget: 17000
  },
  { id: 3003, name: 'APAC Route', owner: 'Kira Long', status: 'Active', startDate: '2026-02-03', dueDate: '2026-08-14', budget: 29000 },
  {
    id: 3004,
    name: 'Legacy Route Cleanup',
    owner: 'Owen Reed',
    status: 'Completed',
    startDate: '2025-09-17',
    dueDate: '2026-01-20',
    budget: 11000
  },
  {
    id: 3005,
    name: 'High-Risk Route',
    owner: 'Diana Ross',
    status: 'Planned',
    startDate: '2026-06-11',
    dueDate: '2026-11-01',
    budget: 22000
  }
];
