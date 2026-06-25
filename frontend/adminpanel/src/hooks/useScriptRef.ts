import { useEffect, useRef } from 'react';

// ==============================|| ELEMENT REFERENCE HOOKS ||============================== //

export default function useScriptRef() {
  const scripted = useRef(true);

  // True while mounted; the cleanup flips it to false on unmount so async
  // handlers can skip state updates after the component is gone.
  useEffect(
    () => () => {
      scripted.current = false;
    },
    []
  );

  return scripted;
}
