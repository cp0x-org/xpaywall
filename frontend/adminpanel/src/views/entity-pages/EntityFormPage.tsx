import { useLocation } from 'react-router-dom';

// material-ui
import Box from '@mui/material/Box';

// project imports
import MainCard from 'ui-component/cards/MainCard';

function getEntityName(pathname: string) {
  if (pathname.startsWith('/payment-channels')) return 'Payment Channel';
  if (pathname.startsWith('/routes')) return 'Route';
  if (pathname.startsWith('/stats')) return 'Stat';
  return 'Entity';
}

export default function EntityFormPage() {
  const { pathname } = useLocation();
  const entityName = getEntityName(pathname);

  let title = entityName;
  if (pathname.includes('/create')) title = `Create ${entityName}`;
  if (pathname.includes('/edit')) title = `Update ${entityName}`;
  if (pathname.includes('/view')) title = `View ${entityName}`;

  return (
    <MainCard title={title}>
      <Box sx={{ minHeight: 320 }} />
    </MainCard>
  );
}
