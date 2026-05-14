import { Activity } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Avatar from '@mui/material/Avatar';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

// project imports
import LogoSection from '../LogoSection';
import ProfileSection from './ProfileSection';

import { handlerDrawerOpen, useGetMenuMaster } from 'api/menu';
import { MenuOrientation } from 'config';
import useConfig from 'hooks/useConfig';

// assets
import { IconMenu2 } from '@tabler/icons-react';

// ==============================|| MAIN NAVBAR / HEADER ||============================== //

export default function Header() {
  const theme = useTheme();
  const downMD = useMediaQuery(theme.breakpoints.down('md'));

  const {
    state: { menuOrientation }
  } = useConfig();
  const { menuMaster } = useGetMenuMaster();
  const drawerOpen = menuMaster.isDashboardDrawerOpened;
  const isHorizontal = menuOrientation === MenuOrientation.HORIZONTAL && !downMD;

  return (
    <>
      {/* logo & toggler button */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <Box component="span" sx={{ display: { xs: 'none', md: 'block' } }}>
          <LogoSection />
        </Box>
        <Activity mode={!isHorizontal ? 'visible' : 'hidden'}>
          <Avatar
            variant="rounded"
            sx={{
              ...theme.typography.commonAvatar,
              ...theme.typography.mediumAvatar,
              overflow: 'hidden',
              transition: 'all .2s ease-in-out',
              color: theme.vars.palette.secondary.dark,
              background: theme.vars.palette.secondary.light,
              '&:hover': {
                color: theme.vars.palette.secondary.light,
                background: theme.vars.palette.secondary.dark
              },
              ...theme.applyStyles('dark', {
                color: theme.vars.palette.secondary.main,
                background: theme.vars.palette.dark.main,
                '&:hover': {
                  color: theme.vars.palette.secondary.light,
                  background: theme.vars.palette.secondary.main
                }
              })
            }}
            onClick={() => handlerDrawerOpen(!drawerOpen)}
          >
            <IconMenu2 stroke={1.5} size="20px" />
          </Avatar>
        </Activity>
      </Box>

      <Box sx={{ flexGrow: 1 }} />
      {/* profile */}
      <ProfileSection />
    </>
  );
}
