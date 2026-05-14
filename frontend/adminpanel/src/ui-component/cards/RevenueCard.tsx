import { ElementType } from 'react';

// material-ui
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';

// ==============================|| REVENUE CARD ||============================== //

interface RevenueCardProps {
  primary: string;
  secondary: string;
  content: string;
  iconPrimary: ElementType;
  color: string;
}

export default function RevenueCard({ primary, secondary, content, iconPrimary: IconPrimary, color }: RevenueCardProps) {
  return (
    <Card sx={{ bgcolor: color, color: 'white', position: 'relative', overflow: 'hidden' }}>
      <CardContent>
        <Box
          sx={{
            position: 'absolute',
            right: -24,
            top: -24,
            opacity: 0.15,
            '& svg': { width: 120, height: 120 }
          }}
        >
          <IconPrimary fontSize="inherit" />
        </Box>
        <Grid container direction="column" spacing={1}>
          <Grid>
            <Typography variant="subtitle2" sx={{ color: 'rgba(255,255,255,0.8)' }}>
              {primary}
            </Typography>
          </Grid>
          <Grid>
            <Typography variant="h3" sx={{ color: 'white' }}>
              {secondary}
            </Typography>
          </Grid>
          <Grid>
            <Typography variant="subtitle2" sx={{ color: 'rgba(255,255,255,0.8)' }}>
              {content}
            </Typography>
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );
}
