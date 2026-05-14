// material-ui
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';

// ==============================|| FOOTER - AUTHENTICATION 2 & 3 ||============================== //

export default function AuthFooter() {
  return (
    <Stack direction="row" sx={{ justifyContent: 'space-between' }}>
      <Typography variant="subtitle2">xpaywall</Typography>
      <Typography variant="subtitle2">&copy; {new Date().getFullYear()} xpaywall</Typography>
    </Stack>
  );
}
