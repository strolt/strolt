export const getTagKey = (tag: string) => {
  let seed = tag;

  const one = tag.split(":");
  if (one.length === 2 && one[0] !== undefined) {
    seed = one[0];
  }

  const two = tag.split("=");
  if (two.length === 2 && two[0] !== undefined) {
    seed = two[0];
  }

  return seed;
};
