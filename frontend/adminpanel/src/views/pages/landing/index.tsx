// material-ui
import { useTheme } from '@mui/material/styles';
import Box from '@mui/material/Box';

// project imports
import AppBar from 'ui-component/extended/AppBar';
import HeaderSection from './HeaderSection';
import CardSection from './CardSection';
import FeatureSection from './FeatureSection';
import HowItWorksSection from './HowItWorksSection';
import FooterSection from './FooterSection';

// =============================|| LANDING MAIN ||============================= //

export default function Landing() {
  const theme = useTheme();

  return (
    <>
      <Box
        id="home"
        sx={{
          overflowX: 'hidden',
          overflowY: 'clip',
          background: `linear-gradient(360deg, ${theme.vars.palette.grey[100]} 1.09%, ${theme.vars.palette.background.paper} 100%)`,
          ...theme.applyStyles('dark', { background: theme.vars.palette.background.default })
        }}
      >
        <AppBar />
        <HeaderSection />
      </Box>

      <Box sx={{ py: 12.5, bgcolor: 'background.default', ...theme.applyStyles('dark', { bgcolor: 'dark.dark' }) }}>
        <CardSection />
      </Box>

      <Box sx={{ py: 12.5, bgcolor: 'grey.100', ...theme.applyStyles('dark', { bgcolor: 'background.default' }) }}>
        <HowItWorksSection />
      </Box>

      <Box sx={{ py: 12.5, bgcolor: 'background.default', ...theme.applyStyles('dark', { bgcolor: 'dark.dark' }) }}>
        <FeatureSection />
      </Box>

      <Box sx={{ py: 12.5, bgcolor: 'dark.900', ...theme.applyStyles('dark', { bgcolor: 'background.default' }), pb: 0 }}>
        <FooterSection />
      </Box>
    </>
  );
}
