import { useCallback, useEffect, useState } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import Tooltip from '@mui/material/Tooltip';
import Box from '@mui/material/Box';

// assets
import { IconArrowsMaximize, IconArrowsMinimize } from '@tabler/icons-react';

// ==============================|| HEADER CONTENT - FULLSCREEN ||============================== //

export default function FullScreen() {
  const theme = useTheme();

  const [open, setOpen] = useState(false);
  const handleToggle = useCallback(() => {
    if (document && !document.fullscreenElement) {
      document.documentElement.requestFullscreen();
    } else if (document.exitFullscreen) {
      document.exitFullscreen();
    }
  }, []);

  useEffect(() => {
    const handleFullscreenChange = () => {
      setOpen(!!document.fullscreenElement);
    };
    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
    };
  }, []);

  return (
    <Box sx={{ ml: 2 }}>
      <Tooltip title={open ? 'Exit Fullscreen' : 'Fullscreen'}>
        <Avatar
          variant="rounded"
          sx={{
            ...theme.typography.commonAvatar,
            ...theme.typography.mediumAvatar,
            transition: 'all .2s ease-in-out',
            color: theme.vars.palette.primary.dark,
            background: theme.vars.palette.primary.light,
            '&:hover': {
              color: theme.vars.palette.primary.light,
              background: theme.vars.palette.primary.main
            },

            ...theme.applyStyles('dark', {
              color: theme.vars.palette.primary.dark,
              background: theme.vars.palette.dark.main,
              '&:hover': {
                color: theme.vars.palette.primary.light,
                background: theme.vars.palette.primary.main
              }
            })
          }}
          aria-controls={open ? 'menu-list-grow' : undefined}
          aria-haspopup="true"
          onClick={handleToggle}
        >
          {open ? <IconArrowsMinimize /> : <IconArrowsMaximize />}
        </Avatar>
      </Tooltip>
    </Box>
  );
}
