export const getTagKey = (tag: string) => {
  let seed = tag;

  const one = tag.split(":");
  if (one.length === 2) {
    seed = one[0];
  }

  const two = tag.split("=");
  if (two.length === 2) {
    seed = two[0];
  }

  return seed;
};
