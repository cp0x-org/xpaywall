// material-ui
import { Theme, useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Button from '@mui/material/Button';
import CardMedia from '@mui/material/CardMedia';
import Stack from '@mui/material/Stack';

// project imports
import useAuth from 'hooks/useAuth';
import { withAlpha } from 'utils/colorUtils';

// assets
import Google from 'assets/images/icons/google.svg';
import Twitter from 'assets/images/icons/twitter.svg';
import Facebook from 'assets/images/icons/facebook.svg';

// ==============================|| FIREBASE - SOCIAL BUTTON ||============================== //

export default function FirebaseSocial() {
  const theme = useTheme();
  const downSM = useMediaQuery((theme: Theme) => theme.breakpoints.down('sm'));

  // @ts-ignore
  const { firebaseFacebookSignIn, firebaseGoogleSignIn, firebaseTwitterSignIn } = useAuth();
  const googleHandler = async () => {
    try {
      await firebaseGoogleSignIn();
    } catch (err) {
      console.error(err);
    }
  };

  const twitterHandler = async () => {
    try {
      await firebaseTwitterSignIn();
    } catch (err) {
      console.error(err);
    }
  };

  const facebookHandler = async () => {
    try {
      await firebaseFacebookSignIn();
    } catch (err) {
      console.error(err);
    }
  };

  const buttonSx = {
    color: 'grey.700',
    bgcolor: 'grey.50',
    borderColor: 'grey.100',
    ...theme.applyStyles('dark', { bgcolor: 'dark.main', borderColor: withAlpha(theme.vars.palette.dark.light, 0.2) })
  };

  return (
    <Stack
      direction="row"
      sx={{ justifyContent: { xs: 'space-around', sm: 'space-between' }, '& .MuiButton-startIcon': { mr: 0 }, gap: { xs: 1, sm: 2 } }}
    >
      <Button
        variant="outlined"
        color="secondary"
        fullWidth={!downSM}
        startIcon={<CardMedia component="img" src={Google} alt="Google" />}
        onClick={googleHandler}
        sx={buttonSx}
      />
      <Button
        variant="outlined"
        color="secondary"
        fullWidth={!downSM}
        startIcon={<CardMedia component="img" src={Twitter} alt="Twitter" />}
        onClick={twitterHandler}
        sx={buttonSx}
      />
      <Button
        variant="outlined"
        color="secondary"
        fullWidth={!downSM}
        startIcon={<CardMedia component="img" src={Facebook} alt="Facebook" />}
        onClick={facebookHandler}
        sx={buttonSx}
      />
    </Stack>
  );
}
