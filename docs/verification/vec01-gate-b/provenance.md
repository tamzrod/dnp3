# Pinned provenance — OpenDNP3 3.1.2 lab peer

All values below are the recorded pinned expectations from
[`../../active_work/vec-01-opendnp3-lab-deployment-readiness.md`](../../active_work/vec-01-opendnp3-lab-deployment-readiness.md)
and the master evidence [`../VEC-01-LAB-EXEC-001.md`](../VEC-01-LAB-EXEC-001.md) §2.1/§2.2,
re-verified against freshly retrieved bytes in §30.2.

## Upstream

| Field | Value |
|-------|-------|
| Upstream | `https://github.com/dnp3/opendnp3` (formerly `automatak/dnp3`) |
| Annotated tag `3.1.2` | `523f7d2d6c65aa671f594fc63c3adcd74e81407d` |
| Peeled commit `3.1.2^{}` | `26b4c01e4839bbbda8866655e086471c4917ee53` |
| Source acquisition URL | `https://codeload.github.com/dnp3/opendnp3/tar.gz/refs/tags/3.1.2` |
| Source tarball SHA-256 | `183cc29222c3cb58099a9753c43defed5cb8677fda985c3f88e38b5c19a36ff2` |

Verify the tag/peel refs live:

```sh
git ls-remote https://github.com/dnp3/opendnp3 'refs/tags/3.1.2*'
# expect: 3.1.2 -> 523f7d2d...   and   3.1.2^{} -> 26b4c01e...
```

## Dependencies (declared `URL_HASH` SHA-1 in `deps/*.cmake` at `26b4c01`)

| Dependency | Pinned version | Upstream URL | Declared SHA-1 | Expected SHA-256 |
|------------|----------------|--------------|----------------|------------------|
| asio | `asio-1-16-0` | `https://github.com/chriskohlhoff/asio/archive/asio-1-16-0.zip` | `6bdd33522d5b95b36445abb2072a481f7ce15402` | `18883c350558eadf91218ae5ce9a8f8f85099c2f036c1c8579fb639092e928c5` |
| exe4cpp | `fb878a4de598ba9d6e4338afebf83f96e03af1b8` | `https://github.com/automatak/exe4cpp/archive/fb878a4de598ba9d6e4338afebf83f96e03af1b8.zip` | `18b141e8cf09dc8d28cc62dd5fa2920670d501bd` | `2a8a2c7d9982f54858b92c03cff91ef433907472039f8e61ff7c9f9b3d48db6c` |
| ser4cpp | `3c449734dc530a8f465eb0982de29165cc4e23d5` | `https://github.com/automatak/ser4cpp/archive/3c449734dc530a8f465eb0982de29165cc4e23d5.zip` | `937b759b7cc80180da26b47037e796b59798a672` | `5a78f83de2d70bdf4be2ab607bc4416c6866ce64a257a8f19948cf5746762f6e` |
| catch (test-only, unused) | `v2.11.3` | `https://raw.githubusercontent.com/catchorg/Catch2/v2.11.3/single_include/catch2/catch.hpp` | `c127ebb7a4f65c6ceff7587c8ef18f84a74d6c15` | `197640c056d6d01e5a56737fbdfe671d2b872c8c7533e01cc2de4be19ab7a028` |

## Expected build artifacts (rebuild expectations, not retained bytes)

| Artifact | Stock (pinned) SHA-256 | Lab (B3, address+variation patched) SHA-256 |
|----------|------------------------|--------------------------------------------|
| `outstation-demo` | `8ae8c84125c4920352b6daef4145e2a91de9e02cd6e2c1b94bf6995ebeb093bb` | `6657ec806f06b72ccbc0fcfcd646ee58d48156fb71b65cba36c59b5b130da3e7` |
| `libopendnp3.a` | `e2926474288fb8d2c6c7962192bad2ff0f724f63c07d2209fe5b30c26fbe0677` | `25f83aaad27fb54b46e98ab215cfe7af9e1fa6f022fc63290e81a564126e03b0` |

**Path-dependence (documented, not a blocker):** the Release binary embeds the absolute
source path in assertion/filename strings. A rebuild at a *different* absolute path yields
a different hash that differs **only** by that embedded path string (of 215 library object
files, 36 differ and all 36 differ solely in the path string; 0 unexplained differences).
Reproducing the historical absolute layout reproduces the pinned stock hashes exactly.
Never claim a modified/lab binary equals the pinned stock binary.

## Licenses

- OpenDNP3: Apache-2.0
- exe4cpp / ser4cpp: BSD-3-Clause
- asio: BSL-1.0

These are rebuildable inputs. No third-party archive or compiled binary is committed to
Git; the pinned URLs and digests above plus `rebuild-and-lab.md` make the peer
reconstructable without retained bytes.
