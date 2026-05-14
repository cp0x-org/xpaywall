// material-ui
import { useTheme } from '@mui/material/styles';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';

// project imports
import FadeInWhenVisible from './Animation';
import SubCard from 'ui-component/cards/SubCard';
import Avatar from 'ui-component/extended/Avatar';

// assets
import RouterIcon from '@mui/icons-material/Router';
import SettingsInputComponentIcon from '@mui/icons-material/SettingsInputComponent';
import BarChartIcon from '@mui/icons-material/BarChart';

const landingCards = [
  {
    title: 'Reverse Proxy',
    caption: 'xgateway transparently proxies requests to your upstream APIs after payment is verified',
    icon: <RouterIcon sx={{ fontSize: '2.25rem' }} />,
    bgcolor: 'primary.200',
    color: 'primary.main'
  },
  {
    title: 'Payment Rules',
    caption: 'Define per-route pricing, payment methods, and merchant addresses via the control API',
    icon: <SettingsInputComponentIcon sx={{ fontSize: '2.25rem' }} />,
    bgcolor: 'warning.main',
    color: 'warning.dark'
  },
  {
    title: 'Real-time Stats',
    caption: 'Track request volumes, payment proofs, and revenue across all your projects and routes',
    icon: <BarChartIcon sx={{ fontSize: '2.25rem' }} />,
    bgcolor: 'secondary.200',
    color: 'secondary.main'
  }
];

// =============================|| LANDING - CARD SECTION ||============================= //

export default function CardSection() {
  const theme = useTheme();

  return (
    <Container>
      <Grid container spacing={{ xs: 3, sm: 5 }} sx={{ justifyContent: 'center', textAlign: 'center' }}>
        {landingCards.map((card, index) => (
          <Grid key={index} size={{ md: 4, sm: 6, xs: 12 }}>
            <FadeInWhenVisible>
              <SubCard
                sx={(theme) => ({
                  bgcolor: card.bgcolor,
                  overflow: 'hidden',
                  position: 'relative',
                  border: 'none',
                  height: 1,
                  '&:after': {
                    content: '""',
                    position: 'absolute',
                    width: 150,
                    height: 150,
                    border: '35px solid',
                    borderColor: 'background.paper',
                    opacity: 0.4,
                    ...theme.applyStyles('dark', { opacity: 0.1 }),
                    borderRadius: '50%',
                    top: -72,
                    right: -63
                  },
                  '&:before': {
                    content: '""',
                    position: 'absolute',
                    width: 150,
                    height: 150,
                    border: '2px solid',
                    borderColor: 'background.paper',
                    opacity: 0.21,
                    ...theme.applyStyles('dark', { opacity: 0.05 }),
                    borderRadius: '50%',
                    top: -97,
                    right: -3
                  },
                  '& .MuiCardContent-root': { padding: '20px 38px 20px 30px' }
                })}
              >
                <Stack sx={{ gap: 2, textAlign: 'left' }}>
                  <Avatar
                    variant="rounded"
                    sx={{
                      bgcolor: 'background.paper',
                      opacity: 0.5,
                      ...theme.applyStyles('dark', { opacity: 1 }),
                      color: card.color,
                      height: 60,
                      width: 60,
                      borderRadius: '12px'
                    }}
                  >
                    {card.icon}
                  </Avatar>
                  <Stack>
                    <Typography variant="h3" sx={{ fontWeight: 700, color: 'grey.900', ...theme.applyStyles('dark', { color: 'dark.900' }) }}>
                      {card.title}
                    </Typography>
                    <Typography variant="body2" sx={{ mt: 1, color: 'grey.800', ...theme.applyStyles('dark', { color: 'dark.800' }) }}>
                      {card.caption}
                    </Typography>
                  </Stack>
                </Stack>
              </SubCard>
            </FadeInWhenVisible>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
}
