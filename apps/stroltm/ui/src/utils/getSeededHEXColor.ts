// const light = '89ABCDEF'
const light = "89abcd00";
// const dark = '01234567'
const dark = "01453d";
const random = dark + light;

type Seed = number | string;

const xmur3 = (str: string): number => {
  let h = 1_779_033_703 ^ str.length;

  for (let i = 0; i < str.length; i++) {
    ((h = Math.imul(h ^ str.codePointAt(i), 3_432_918_353)), (h = (h << 13) | (h >>> 19)));
  }
  h = Math.imul(h ^ (h >>> 16), 2_246_822_507);
  h = Math.imul(h ^ (h >>> 13), 3_266_489_909);
  return (h ^= h >>> 16) >>> 0;
};

const _randomColor = (list: string, seed?: Seed): string => {
  const seedNumber = seed
    ? xmur3(seed.toString())
    : Number((Math.random() * Date.now()).toFixed(0));
  const fixedSeedNumber = seedNumber % 1e6;
  const arrayOfNumber = [...fixedSeedNumber.toString()].map((n) => Number(n));

  while (arrayOfNumber.length < 6) {
    arrayOfNumber.push(Number(((seedNumber / arrayOfNumber.length) % 10).toFixed(0)));
  }

  const isRandom = list.length === random.length;
  const indexColor = arrayOfNumber.map((n, i) => {
    const isOdd = i % 2;
    if (isRandom || isOdd) {
      let m = 0;
      m = n * 2;
      if (isOdd) {
        m += 1;
      }
      if (m < random.length) {
        return m;
      }
      const o = (m - random.length) * 3 + 1;
      if (isOdd && o % 2) {
        return o + 1;
      }
      if (isOdd) {
        return o - 1;
      }
      return o;
    }
    if (n < list.length) {
      return n;
    }
    return [4, 5, 6, 7, 8, 9].indexOf(n - i);
  });

  let hexResult = "";

  for (let i = 0; i < indexColor.length; i++) {
    const n = indexColor[i];
    hexResult += (i % 2 ? random[n] : list[n]) || "a";
  }
  return hexResult;
};

const randomLightColor = (seed?: Seed): string => _randomColor(light, seed);

const randomDarkColor = (seed?: Seed): string => _randomColor(dark, seed);

export const getSeededHEXColor = (seed?: Seed, mode?: "dark" | "light"): string => {
  let c = "";
  if (mode === "dark") {
    c = randomLightColor(seed);
  } else {
    c = randomDarkColor(seed);
  }

  return [...`#${c}`].slice(0, 7).join("");
};
