import { ElementType } from 'react';

// material-ui
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';

// ==============================|| SIDE ICON CARD ||============================== //

interface SideIconCardProps {
  iconPrimary: ElementType;
  primary: string;
  secondary: string;
  secondarySub: string;
  color: string;
}

export default function SideIconCard({ iconPrimary: IconPrimary, primary, secondary, secondarySub, color }: SideIconCardProps) {
  return (
    <Card>
      <CardContent>
        <Grid container spacing={0} alignItems="center" justifyContent="space-between">
          <Grid>
            <Box sx={{ color, '& svg': { width: 56, height: 56 } }}>
              <IconPrimary fontSize="inherit" />
            </Box>
          </Grid>
          <Grid>
            <Typography variant="h3" align="right">
              {primary}
            </Typography>
            <Typography variant="subtitle1" align="right">
              {secondary}&nbsp;
              <Typography component="span" variant="subtitle1" sx={{ color }}>
                {secondarySub}
              </Typography>
            </Typography>
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );
}
