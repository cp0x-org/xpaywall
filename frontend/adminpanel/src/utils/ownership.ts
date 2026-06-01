const SUPERADMIN_ID = '00000000-0000-0000-0000-000000000000';

export function canManage(currentUserId: string | undefined | null, ownerUserId: string | undefined | null): boolean {
  if (!currentUserId) return false;
  if (currentUserId === SUPERADMIN_ID) return true;
  if (!ownerUserId) return true;
  return currentUserId === ownerUserId;
}
