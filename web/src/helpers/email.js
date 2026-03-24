export const OPENCUMT_EMAIL_HINT = '@opencumt.org';
export const OPENCUMT_EMAIL_REGEX = /^[A-Za-z0-9._%+-]+@opencumt\.org$/i;

export function normalizeOrgEmail(value = '') {
  return value.trim().toLowerCase();
}

export function isOpenCUMTEmail(value = '') {
  return OPENCUMT_EMAIL_REGEX.test(normalizeOrgEmail(value));
}
