
const DEFAULT_NGRAM_LEN = 3;

/**
 * Low-level Unicode-aware N-gram index.
 */
export class NgramIndex {
  constructor({ ngramLen = DEFAULT_NGRAM_LEN } = {}) {
    if (!Number.isInteger(ngramLen) || ngramLen < 1) {
      throw new RangeError("ngramLen must be a positive integer");
    }

    this.ngramLen = ngramLen;
    this.index = new Map();
    this.docs = [];
  }

  /**
   * Add document associated with one or more texts.
   *
   * @param {*} doc
   * @param {...string} texts
   */
  add(doc, ...texts) {
    const docIndex = this.docs.length;
    let ngramCount = 0;

    for (const text of texts) {
      const chars = Array.from(text);

      for (let i = 0; i <= chars.length - this.ngramLen; i++) {
        const ngram = chars
          .slice(i, i + this.ngramLen)
          .join("");

        let occurrences = this.index.get(ngram);

        if (!occurrences) {
          occurrences = [];
          this.index.set(ngram, occurrences);
        }

        occurrences.push(docIndex);
        ngramCount++;
      }
    }

    // Same behavior as the Go implementation:
    // documents without n-grams are not indexed.
    if (ngramCount > 0) {
      this.docs.push({
        document: doc,
        ngramCount
      });
    }
  }

  /**
   * Return detailed matching information.
   *
   * @param {...string} texts
   * @returns {Array<{
   *   document: *,
   *   totalNgrams: number,
   *   matchedNgrams: number,
   *   matchRate: number
   * }>}
   */
  lookup(...texts) {
    const counts = new Array(this.docs.length).fill(0);

    for (const text of texts) {
      const chars = Array.from(text);

      for (let i = 0; i <= chars.length - this.ngramLen; i++) {
        const ngram = chars
          .slice(i, i + this.ngramLen)
          .join("");

        const occurrences = this.index.get(ngram);

        if (!occurrences) continue;

        for (const docIndex of occurrences) {
          counts[docIndex]++;
        }
      }
    }

    const results = [];

    for (let i = 0; i < counts.length; i++) {
      const matchedNgrams = counts[i];

      if (matchedNgrams === 0) continue;

      const doc = this.docs[i];

      results.push({
        document: doc.document,
        totalNgrams: doc.ngramCount,
        matchedNgrams,
        matchRate: matchedNgrams / doc.ngramCount
      });
    }

    return results;
  }

  /**
   * Return documents sorted by number of matched n-grams.
   */
  search(...texts) {
    return this.lookup(...texts)
      .sort((a, b) => b.matchedNgrams - a.matchedNgrams)
      .map(result => result.document);
  }
}

/**
 * String-based index with customizable normalization.
 */
export class StringNgramIndex {
  constructor({
    ngramLen = DEFAULT_NGRAM_LEN,
    normalizer = text => [text],
    comparator = (a, b) =>
      b.matchedNgrams - a.matchedNgrams
  } = {}) {
    this.index = new NgramIndex({ ngramLen });
    this.normalizer = normalizer;
    this.comparator = comparator;
  }

  /**
   * Add document associated with one or more strings.
   */
  add(doc, ...texts) {
    const normalized = this.normalize(texts);
    this.index.add(doc, ...normalized);
  }

  /**
   * Detailed search results.
   */
  lookup(...texts) {
    return this.index.lookup(...this.normalize(texts));
  }

  /**
   * Search and return documents.
   */
  search(...texts) {
    return this.lookup(...texts)
      .sort(this.comparator)
      .map(result => result.document);
  }

  /**
   * Normalize input strings.
   *
   * Normalizer returns an array of strings.
   * Each string is indexed independently.
   */
  normalize(texts) {
    const result = [];

    for (const text of texts) {
      const normalized = this.normalizer(text);

      if (!Array.isArray(normalized)) {
        throw new TypeError(
          "Normalizer must return an array of strings"
        );
      }

      for (const part of normalized) {
        if (typeof part !== "string") {
          throw new TypeError(
            "Normalized values must be strings"
          );
        }

        result.push(part);
      }
    }

    return result;
  }
}
