import { ElementType } from 'react';

// material-ui
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';

// ==============================|| HOVER DATA CARD ||============================== //

interface HoverDataCardProps {
  title: string;
  primary: number;
  color: string;
}

export default function HoverDataCard({ title, primary }: HoverDataCardProps) {
  return (
    <Card
      sx={{
        transition: 'box-shadow 0.3s',
        '&:hover': { boxShadow: 8 }
      }}
    >
      <CardContent sx={{ textAlign: 'center' }}>
        {/*<IconPrimary sx={{ color, fontSize: 36, mb: 1 }} />*/}
        <Typography variant="subtitle2" color="text.secondary" gutterBottom>
          {title}
        </Typography>
        <Typography variant="h3" sx={{ mb: 1 }}>
          {primary.toLocaleString()}
        </Typography>
      </CardContent>
    </Card>
  );
}
