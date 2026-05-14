import { EntityRow } from 'views/entity-pages/types';

export const PAYMENT_CHANNEL_ROWS: EntityRow[] = [
  { id: 2001, name: 'Stripe Gateway', owner: 'Elena Brooks', status: 'Active', startDate: '2026-01-15', dueDate: '2026-07-01', budget: 26000 },
  { id: 2002, name: 'PayPal Express', owner: 'Liam Carter', status: 'Planned', startDate: '2026-05-12', dueDate: '2026-08-20', budget: 18000 },
  { id: 2003, name: 'SEPA Direct', owner: 'Mila Novak', status: 'Completed', startDate: '2025-10-03', dueDate: '2026-01-30', budget: 12000 },
  { id: 2004, name: 'Crypto Checkout', owner: 'Victor Hall', status: 'Active', startDate: '2026-02-18', dueDate: '2026-09-10', budget: 34000 },
  { id: 2005, name: 'Bank Wire', owner: 'Nora White', status: 'Planned', startDate: '2026-06-05', dueDate: '2026-10-15', budget: 14000 }
];
