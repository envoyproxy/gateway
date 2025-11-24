// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/envoyproxy/gateway/internal/utils/test"
	"github.com/envoyproxy/gateway/test/benchmark/proto"
)

func fakeCaseResult() *proto.Result {
	data := []byte(`{
   "name": "global",
   "statistics": [
    {
     "count": "0",
     "id": "benchmark_http_client.latency_1xx",
     "mean": "0s",
     "pstdev": "0s",
     "percentiles": [
      {
       "duration": "0s",
       "percentile": 1,
       "count": "0"
      }
     ],
     "min": "0s",
     "max": "0s"
    },
    {
     "count": "489025",
     "id": "benchmark_http_client.latency_2xx",
     "mean": "0.001206029s",
     "pstdev": "0.006470055s",
     "percentiles": [
      {
       "duration": "0.000135839s",
       "percentile": 0,
       "count": "1"
      },
      {
       "duration": "0.000320303s",
       "percentile": 0.1,
       "count": "48927"
      },
      {
       "duration": "0.000352383s",
       "percentile": 0.2,
       "count": "97874"
      },
      {
       "duration": "0.000379967s",
       "percentile": 0.3,
       "count": "146775"
      },
      {
       "duration": "0.000407919s",
       "percentile": 0.4,
       "count": "195613"
      },
      {
       "duration": "0.000438127s",
       "percentile": 0.5,
       "count": "244542"
      },
      {
       "duration": "0.000454671s",
       "percentile": 0.55,
       "count": "268970"
      },
      {
       "duration": "0.000472799s",
       "percentile": 0.6,
       "count": "293421"
      },
      {
       "duration": "0.000493135s",
       "percentile": 0.65,
       "count": "317905"
      },
      {
       "duration": "0.000517167s",
       "percentile": 0.7,
       "count": "342357"
      },
      {
       "duration": "0.000546047s",
       "percentile": 0.75,
       "count": "366785"
      },
      {
       "duration": "0.000563551s",
       "percentile": 0.775,
       "count": "378998"
      },
      {
       "duration": "0.000584511s",
       "percentile": 0.8,
       "count": "391239"
      },
      {
       "duration": "0.000609887s",
       "percentile": 0.825,
       "count": "403448"
      },
      {
       "duration": "0.000642047s",
       "percentile": 0.85,
       "count": "415681"
      },
      {
       "duration": "0.000685919s",
       "percentile": 0.875,
       "count": "427910"
      },
      {
       "duration": "0.000715263s",
       "percentile": 0.8875,
       "count": "434012"
      },
      {
       "duration": "0.000751775s",
       "percentile": 0.9,
       "count": "440125"
      },
      {
       "duration": "0.000799647s",
       "percentile": 0.9125,
       "count": "446236"
      },
      {
       "duration": "0.000863807s",
       "percentile": 0.925,
       "count": "452353"
      },
      {
       "duration": "0.000956479s",
       "percentile": 0.9375,
       "count": "458461"
      },
      {
       "duration": "0.001018239s",
       "percentile": 0.94375,
       "count": "461519"
      },
      {
       "duration": "0.001100415s",
       "percentile": 0.95,
       "count": "464574"
      },
      {
       "duration": "0.001211903s",
       "percentile": 0.95625,
       "count": "467631"
      },
      {
       "duration": "0.001372223s",
       "percentile": 0.9625,
       "count": "470687"
      },
      {
       "duration": "0.001629759s",
       "percentile": 0.96875,
       "count": "473743"
      },
      {
       "duration": "0.001838463s",
       "percentile": 0.971875,
       "count": "475272"
      },
      {
       "duration": "0.002123263s",
       "percentile": 0.975,
       "count": "476800"
      },
      {
       "duration": "0.002562687s",
       "percentile": 0.978125,
       "count": "478329"
      },
      {
       "duration": "0.003192703s",
       "percentile": 0.98125,
       "count": "479856"
      },
      {
       "duration": "0.004134399s",
       "percentile": 0.984375,
       "count": "481385"
      },
      {
       "duration": "0.005107455s",
       "percentile": 0.9859375,
       "count": "482149"
      },
      {
       "duration": "0.008261119s",
       "percentile": 0.9875,
       "count": "482913"
      },
      {
       "duration": "0.049223679s",
       "percentile": 0.9890625,
       "count": "483677"
      },
      {
       "duration": "0.055601151s",
       "percentile": 0.990625,
       "count": "484444"
      },
      {
       "duration": "0.057024511s",
       "percentile": 0.9921875,
       "count": "485206"
      },
      {
       "duration": "0.057501695s",
       "percentile": 0.99296875,
       "count": "485587"
      },
      {
       "duration": "0.057905151s",
       "percentile": 0.99375,
       "count": "485969"
      },
      {
       "duration": "0.058208255s",
       "percentile": 0.99453125,
       "count": "486351"
      },
      {
       "duration": "0.058564607s",
       "percentile": 0.9953125,
       "count": "486733"
      },
      {
       "duration": "0.058861567s",
       "percentile": 0.99609375,
       "count": "487115"
      },
      {
       "duration": "0.059035647s",
       "percentile": 0.996484375,
       "count": "487310"
      },
      {
       "duration": "0.059203583s",
       "percentile": 0.996875,
       "count": "487497"
      },
      {
       "duration": "0.059373567s",
       "percentile": 0.997265625,
       "count": "487689"
      },
      {
       "duration": "0.059602943s",
       "percentile": 0.99765625,
       "count": "487880"
      },
      {
       "duration": "0.059774975s",
       "percentile": 0.998046875,
       "count": "488073"
      },
      {
       "duration": "0.059893759s",
       "percentile": 0.9982421875,
       "count": "488166"
      },
      {
       "duration": "0.060014591s",
       "percentile": 0.9984375,
       "count": "488261"
      },
      {
       "duration": "0.060151807s",
       "percentile": 0.9986328125,
       "count": "488357"
      },
      {
       "duration": "0.060289023s",
       "percentile": 0.998828125,
       "count": "488453"
      },
      {
       "duration": "0.060461055s",
       "percentile": 0.9990234375,
       "count": "488548"
      },
      {
       "duration": "0.060547071s",
       "percentile": 0.99912109375,
       "count": "488596"
      },
      {
       "duration": "0.060649471s",
       "percentile": 0.99921875,
       "count": "488643"
      },
      {
       "duration": "0.060764159s",
       "percentile": 0.99931640625,
       "count": "488691"
      },
      {
       "duration": "0.060907519s",
       "percentile": 0.9994140625,
       "count": "488739"
      },
      {
       "duration": "0.061030399s",
       "percentile": 0.99951171875,
       "count": "488787"
      },
      {
       "duration": "0.061085695s",
       "percentile": 0.999560546875,
       "count": "488811"
      },
      {
       "duration": "0.061153279s",
       "percentile": 0.999609375,
       "count": "488834"
      },
      {
       "duration": "0.061229055s",
       "percentile": 0.999658203125,
       "count": "488858"
      },
      {
       "duration": "0.061327359s",
       "percentile": 0.99970703125,
       "count": "488882"
      },
      {
       "duration": "0.061523967s",
       "percentile": 0.999755859375,
       "count": "488906"
      },
      {
       "duration": "0.061640703s",
       "percentile": 0.9997802734375,
       "count": "488918"
      },
      {
       "duration": "0.061755391s",
       "percentile": 0.9998046875,
       "count": "488930"
      },
      {
       "duration": "0.061833215s",
       "percentile": 0.9998291015625,
       "count": "488942"
      },
      {
       "duration": "0.062103551s",
       "percentile": 0.999853515625,
       "count": "488954"
      },
      {
       "duration": "0.062275583s",
       "percentile": 0.9998779296875,
       "count": "488966"
      },
      {
       "duration": "0.062425087s",
       "percentile": 0.99989013671875,
       "count": "488972"
      },
      {
       "duration": "0.062658559s",
       "percentile": 0.99990234375,
       "count": "488978"
      },
      {
       "duration": "0.063107071s",
       "percentile": 0.99991455078125,
       "count": "488984"
      },
      {
       "duration": "0.063416319s",
       "percentile": 0.9999267578125,
       "count": "488990"
      },
      {
       "duration": "0.063924223s",
       "percentile": 0.99993896484375,
       "count": "488996"
      },
      {
       "duration": "0.067158015s",
       "percentile": 0.999945068359375,
       "count": "488999"
      },
      {
       "duration": "0.067633151s",
       "percentile": 0.999951171875,
       "count": "489002"
      },
      {
       "duration": "0.070033407s",
       "percentile": 0.999957275390625,
       "count": "489005"
      },
      {
       "duration": "0.076083199s",
       "percentile": 0.99996337890625,
       "count": "489008"
      },
      {
       "duration": "0.083476479s",
       "percentile": 0.999969482421875,
       "count": "489011"
      },
      {
       "duration": "0.123727871s",
       "percentile": 0.99997253417968746,
       "count": "489012"
      },
      {
       "duration": "0.124600319s",
       "percentile": 0.9999755859375,
       "count": "489014"
      },
      {
       "duration": "0.124698623s",
       "percentile": 0.99997863769531248,
       "count": "489015"
      },
      {
       "duration": "0.473071615s",
       "percentile": 0.999981689453125,
       "count": "489018"
      },
      {
       "duration": "0.473071615s",
       "percentile": 0.9999847412109375,
       "count": "489018"
      },
      {
       "duration": "0.473235455s",
       "percentile": 0.99998626708984373,
       "count": "489019"
      },
      {
       "duration": "0.473268223s",
       "percentile": 0.99998779296875,
       "count": "489020"
      },
      {
       "duration": "0.473268223s",
       "percentile": 0.99998931884765629,
       "count": "489020"
      },
      {
       "duration": "0.473530367s",
       "percentile": 0.99999084472656252,
       "count": "489021"
      },
      {
       "duration": "0.473546751s",
       "percentile": 0.99999237060546875,
       "count": "489022"
      },
      {
       "duration": "0.473546751s",
       "percentile": 0.99999313354492192,
       "count": "489022"
      },
      {
       "duration": "0.473645055s",
       "percentile": 0.999993896484375,
       "count": "489023"
      },
      {
       "duration": "0.473645055s",
       "percentile": 0.99999465942382815,
       "count": "489023"
      },
      {
       "duration": "0.473645055s",
       "percentile": 0.99999542236328121,
       "count": "489023"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.99999618530273438,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.999996566772461,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.99999694824218754,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.999997329711914,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.9999977111816406,
       "count": "489024"
      },
      {
       "duration": "0.473694207s",
       "percentile": 0.99999809265136719,
       "count": "489025"
      },
      {
       "duration": "0.473694207s",
       "percentile": 1,
       "count": "489025"
      }
     ],
     "min": "0.000135832s",
     "max": "0.473694207s"
    },
    {
     "count": "0",
     "id": "benchmark_http_client.latency_3xx",
     "mean": "0s",
     "pstdev": "0s",
     "percentiles": [
      {
       "duration": "0s",
       "percentile": 1,
       "count": "0"
      }
     ],
     "min": "0s",
     "max": "0s"
    },
    {
     "count": "0",
     "id": "benchmark_http_client.latency_4xx",
     "mean": "0s",
     "pstdev": "0s",
     "percentiles": [
      {
       "duration": "0s",
       "percentile": 1,
       "count": "0"
      }
     ],
     "min": "0s",
     "max": "0s"
    },
    {
     "count": "0",
     "id": "benchmark_http_client.latency_5xx",
     "mean": "0s",
     "pstdev": "0s",
     "percentiles": [
      {
       "duration": "0s",
       "percentile": 1,
       "count": "0"
      }
     ],
     "min": "0s",
     "max": "0s"
    },
    {
     "count": "0",
     "id": "benchmark_http_client.latency_xxx",
     "mean": "0s",
     "pstdev": "0s",
     "percentiles": [
      {
       "duration": "0s",
       "percentile": 1,
       "count": "0"
      }
     ],
     "min": "0s",
     "max": "0s"
    },
    {
     "count": "0",
     "id": "benchmark_http_client.origin_latency_statistic",
     "mean": "0s",
     "pstdev": "0s",
     "percentiles": [
      {
       "duration": "0s",
       "percentile": 1,
       "count": "0"
      }
     ],
     "min": "0s",
     "max": "0s"
    },
    {
     "count": "489026",
     "id": "benchmark_http_client.queue_to_connect",
     "mean": "0.000010746s",
     "pstdev": "0.000822232s",
     "percentiles": [
      {
       "duration": "0.000002250s",
       "percentile": 0,
       "count": "1"
      },
      {
       "duration": "0.000004s",
       "percentile": 0.1,
       "count": "50366"
      },
      {
       "duration": "0.000004500s",
       "percentile": 0.2,
       "count": "100362"
      },
      {
       "duration": "0.000004875s",
       "percentile": 0.3,
       "count": "148530"
      },
      {
       "duration": "0.000005250s",
       "percentile": 0.4,
       "count": "195813"
      },
      {
       "duration": "0.000005666s",
       "percentile": 0.5,
       "count": "245388"
      },
      {
       "duration": "0.000005833s",
       "percentile": 0.55,
       "count": "269232"
      },
      {
       "duration": "0.000006s",
       "percentile": 0.6,
       "count": "293480"
      },
      {
       "duration": "0.000006208s",
       "percentile": 0.65,
       "count": "319152"
      },
      {
       "duration": "0.000006416s",
       "percentile": 0.7,
       "count": "343865"
      },
      {
       "duration": "0.000006584s",
       "percentile": 0.75,
       "count": "367127"
      },
      {
       "duration": "0.000006708s",
       "percentile": 0.775,
       "count": "379714"
      },
      {
       "duration": "0.000006833s",
       "percentile": 0.8,
       "count": "392841"
      },
      {
       "duration": "0.000006958s",
       "percentile": 0.825,
       "count": "404808"
      },
      {
       "duration": "0.000007084s",
       "percentile": 0.85,
       "count": "416352"
      },
      {
       "duration": "0.000007291s",
       "percentile": 0.875,
       "count": "428276"
      },
      {
       "duration": "0.000007375s",
       "percentile": 0.8875,
       "count": "434094"
      },
      {
       "duration": "0.000007542s",
       "percentile": 0.9,
       "count": "440995"
      },
      {
       "duration": "0.000007709s",
       "percentile": 0.9125,
       "count": "446371"
      },
      {
       "duration": "0.000007958s",
       "percentile": 0.925,
       "count": "452354"
      },
      {
       "duration": "0.000008292s",
       "percentile": 0.9375,
       "count": "458532"
      },
      {
       "duration": "0.000008583s",
       "percentile": 0.94375,
       "count": "461757"
      },
      {
       "duration": "0.000008958s",
       "percentile": 0.95,
       "count": "464743"
      },
      {
       "duration": "0.000009459s",
       "percentile": 0.95625,
       "count": "467643"
      },
      {
       "duration": "0.000010250s",
       "percentile": 0.9625,
       "count": "470745"
      },
      {
       "duration": "0.000011667s",
       "percentile": 0.96875,
       "count": "473746"
      },
      {
       "duration": "0.000012833s",
       "percentile": 0.971875,
       "count": "475273"
      },
      {
       "duration": "0.000014583s",
       "percentile": 0.975,
       "count": "476812"
      },
      {
       "duration": "0.000017084s",
       "percentile": 0.978125,
       "count": "478330"
      },
      {
       "duration": "0.000020667s",
       "percentile": 0.98125,
       "count": "479868"
      },
      {
       "duration": "0.000025375s",
       "percentile": 0.984375,
       "count": "481390"
      },
      {
       "duration": "0.000028292s",
       "percentile": 0.9859375,
       "count": "482158"
      },
      {
       "duration": "0.000031584s",
       "percentile": 0.9875,
       "count": "482917"
      },
      {
       "duration": "0.000034459s",
       "percentile": 0.9890625,
       "count": "483686"
      },
      {
       "duration": "0.000037459s",
       "percentile": 0.990625,
       "count": "484447"
      },
      {
       "duration": "0.000040917s",
       "percentile": 0.9921875,
       "count": "485214"
      },
      {
       "duration": "0.000042875s",
       "percentile": 0.99296875,
       "count": "485596"
      },
      {
       "duration": "0.000045793s",
       "percentile": 0.99375,
       "count": "485970"
      },
      {
       "duration": "0.000051209s",
       "percentile": 0.99453125,
       "count": "486355"
      },
      {
       "duration": "0.000059459s",
       "percentile": 0.9953125,
       "count": "486736"
      },
      {
       "duration": "0.000071127s",
       "percentile": 0.99609375,
       "count": "487116"
      },
      {
       "duration": "0.000079419s",
       "percentile": 0.996484375,
       "count": "487308"
      },
      {
       "duration": "0.000088583s",
       "percentile": 0.996875,
       "count": "487498"
      },
      {
       "duration": "0.000099751s",
       "percentile": 0.997265625,
       "count": "487689"
      },
      {
       "duration": "0.000111791s",
       "percentile": 0.99765625,
       "count": "487880"
      },
      {
       "duration": "0.000125627s",
       "percentile": 0.998046875,
       "count": "488071"
      },
      {
       "duration": "0.000137631s",
       "percentile": 0.9982421875,
       "count": "488167"
      },
      {
       "duration": "0.000149375s",
       "percentile": 0.9984375,
       "count": "488262"
      },
      {
       "duration": "0.000163087s",
       "percentile": 0.9986328125,
       "count": "488358"
      },
      {
       "duration": "0.000178711s",
       "percentile": 0.998828125,
       "count": "488453"
      },
      {
       "duration": "0.000206335s",
       "percentile": 0.9990234375,
       "count": "488549"
      },
      {
       "duration": "0.000217671s",
       "percentile": 0.99912109375,
       "count": "488597"
      },
      {
       "duration": "0.000231799s",
       "percentile": 0.99921875,
       "count": "488644"
      },
      {
       "duration": "0.000256375s",
       "percentile": 0.99931640625,
       "count": "488692"
      },
      {
       "duration": "0.000281007s",
       "percentile": 0.9994140625,
       "count": "488740"
      },
      {
       "duration": "0.000314431s",
       "percentile": 0.99951171875,
       "count": "488788"
      },
      {
       "duration": "0.000329263s",
       "percentile": 0.999560546875,
       "count": "488812"
      },
      {
       "duration": "0.000356255s",
       "percentile": 0.999609375,
       "count": "488835"
      },
      {
       "duration": "0.000395919s",
       "percentile": 0.999658203125,
       "count": "488859"
      },
      {
       "duration": "0.000460095s",
       "percentile": 0.99970703125,
       "count": "488883"
      },
      {
       "duration": "0.000544383s",
       "percentile": 0.999755859375,
       "count": "488907"
      },
      {
       "duration": "0.000565151s",
       "percentile": 0.9997802734375,
       "count": "488919"
      },
      {
       "duration": "0.000635679s",
       "percentile": 0.9998046875,
       "count": "488931"
      },
      {
       "duration": "0.000700543s",
       "percentile": 0.9998291015625,
       "count": "488943"
      },
      {
       "duration": "0.000773151s",
       "percentile": 0.999853515625,
       "count": "488955"
      },
      {
       "duration": "0.000958719s",
       "percentile": 0.9998779296875,
       "count": "488967"
      },
      {
       "duration": "0.001070207s",
       "percentile": 0.99989013671875,
       "count": "488973"
      },
      {
       "duration": "0.001166271s",
       "percentile": 0.99990234375,
       "count": "488979"
      },
      {
       "duration": "0.001289023s",
       "percentile": 0.99991455078125,
       "count": "488985"
      },
      {
       "duration": "0.001406015s",
       "percentile": 0.9999267578125,
       "count": "488991"
      },
      {
       "duration": "0.001588671s",
       "percentile": 0.99993896484375,
       "count": "488997"
      },
      {
       "duration": "0.001754175s",
       "percentile": 0.999945068359375,
       "count": "489000"
      },
      {
       "duration": "0.002260351s",
       "percentile": 0.999951171875,
       "count": "489003"
      },
      {
       "duration": "0.002383103s",
       "percentile": 0.999957275390625,
       "count": "489006"
      },
      {
       "duration": "0.003539199s",
       "percentile": 0.99996337890625,
       "count": "489009"
      },
      {
       "duration": "0.004049919s",
       "percentile": 0.999969482421875,
       "count": "489012"
      },
      {
       "duration": "0.004100479s",
       "percentile": 0.99997253417968746,
       "count": "489013"
      },
      {
       "duration": "0.005742335s",
       "percentile": 0.9999755859375,
       "count": "489015"
      },
      {
       "duration": "0.009892863s",
       "percentile": 0.99997863769531248,
       "count": "489016"
      },
      {
       "duration": "0.181469183s",
       "percentile": 0.999981689453125,
       "count": "489018"
      },
      {
       "duration": "0.181690367s",
       "percentile": 0.9999847412109375,
       "count": "489019"
      },
      {
       "duration": "0.181698559s",
       "percentile": 0.99998626708984373,
       "count": "489020"
      },
      {
       "duration": "0.181772287s",
       "percentile": 0.99998779296875,
       "count": "489021"
      },
      {
       "duration": "0.181772287s",
       "percentile": 0.99998931884765629,
       "count": "489021"
      },
      {
       "duration": "0.181780479s",
       "percentile": 0.99999084472656252,
       "count": "489023"
      },
      {
       "duration": "0.181780479s",
       "percentile": 0.99999237060546875,
       "count": "489023"
      },
      {
       "duration": "0.181780479s",
       "percentile": 0.99999313354492192,
       "count": "489023"
      },
      {
       "duration": "0.181788671s",
       "percentile": 0.999993896484375,
       "count": "489024"
      },
      {
       "duration": "0.181788671s",
       "percentile": 0.99999465942382815,
       "count": "489024"
      },
      {
       "duration": "0.181788671s",
       "percentile": 0.99999542236328121,
       "count": "489024"
      },
      {
       "duration": "0.181870591s",
       "percentile": 0.99999618530273438,
       "count": "489025"
      },
      {
       "duration": "0.181870591s",
       "percentile": 0.999996566772461,
       "count": "489025"
      },
      {
       "duration": "0.181870591s",
       "percentile": 0.99999694824218754,
       "count": "489025"
      },
      {
       "duration": "0.181870591s",
       "percentile": 0.999997329711914,
       "count": "489025"
      },
      {
       "duration": "0.181870591s",
       "percentile": 0.9999977111816406,
       "count": "489025"
      },
      {
       "duration": "0.181919743s",
       "percentile": 0.99999809265136719,
       "count": "489026"
      },
      {
       "duration": "0.181919743s",
       "percentile": 1,
       "count": "489026"
      }
     ],
     "min": "0.000002250s",
     "max": "0.181919743s"
    },
    {
     "count": "489025",
     "id": "benchmark_http_client.request_to_response",
     "mean": "0.001205247s",
     "pstdev": "0.006469448s",
     "percentiles": [
      {
       "duration": "0.000135167s",
       "percentile": 0,
       "count": "1"
      },
      {
       "duration": "0.000319887s",
       "percentile": 0.1,
       "count": "48928"
      },
      {
       "duration": "0.000351919s",
       "percentile": 0.2,
       "count": "97833"
      },
      {
       "duration": "0.000379503s",
       "percentile": 0.3,
       "count": "146754"
      },
      {
       "duration": "0.000407471s",
       "percentile": 0.4,
       "count": "195617"
      },
      {
       "duration": "0.000437679s",
       "percentile": 0.5,
       "count": "244566"
      },
      {
       "duration": "0.000454223s",
       "percentile": 0.55,
       "count": "269005"
      },
      {
       "duration": "0.000472335s",
       "percentile": 0.6,
       "count": "293463"
      },
      {
       "duration": "0.000492639s",
       "percentile": 0.65,
       "count": "317913"
      },
      {
       "duration": "0.000516671s",
       "percentile": 0.7,
       "count": "342362"
      },
      {
       "duration": "0.000545567s",
       "percentile": 0.75,
       "count": "366783"
      },
      {
       "duration": "0.000563071s",
       "percentile": 0.775,
       "count": "379015"
      },
      {
       "duration": "0.000583935s",
       "percentile": 0.8,
       "count": "391232"
      },
      {
       "duration": "0.000609343s",
       "percentile": 0.825,
       "count": "403461"
      },
      {
       "duration": "0.000641471s",
       "percentile": 0.85,
       "count": "415678"
      },
      {
       "duration": "0.000685311s",
       "percentile": 0.875,
       "count": "427908"
      },
      {
       "duration": "0.000714591s",
       "percentile": 0.8875,
       "count": "434011"
      },
      {
       "duration": "0.000751071s",
       "percentile": 0.9,
       "count": "440124"
      },
      {
       "duration": "0.000798879s",
       "percentile": 0.9125,
       "count": "446241"
      },
      {
       "duration": "0.000863007s",
       "percentile": 0.925,
       "count": "452352"
      },
      {
       "duration": "0.000955583s",
       "percentile": 0.9375,
       "count": "458462"
      },
      {
       "duration": "0.001017439s",
       "percentile": 0.94375,
       "count": "461520"
      },
      {
       "duration": "0.001099455s",
       "percentile": 0.95,
       "count": "464575"
      },
      {
       "duration": "0.001211007s",
       "percentile": 0.95625,
       "count": "467632"
      },
      {
       "duration": "0.001371263s",
       "percentile": 0.9625,
       "count": "470687"
      },
      {
       "duration": "0.001628351s",
       "percentile": 0.96875,
       "count": "473743"
      },
      {
       "duration": "0.001836927s",
       "percentile": 0.971875,
       "count": "475273"
      },
      {
       "duration": "0.002120831s",
       "percentile": 0.975,
       "count": "476800"
      },
      {
       "duration": "0.002560255s",
       "percentile": 0.978125,
       "count": "478328"
      },
      {
       "duration": "0.003190911s",
       "percentile": 0.98125,
       "count": "479856"
      },
      {
       "duration": "0.004130559s",
       "percentile": 0.984375,
       "count": "481384"
      },
      {
       "duration": "0.005104127s",
       "percentile": 0.9859375,
       "count": "482149"
      },
      {
       "duration": "0.008174591s",
       "percentile": 0.9875,
       "count": "482913"
      },
      {
       "duration": "0.049221631s",
       "percentile": 0.9890625,
       "count": "483677"
      },
      {
       "duration": "0.055599103s",
       "percentile": 0.990625,
       "count": "484442"
      },
      {
       "duration": "0.057022463s",
       "percentile": 0.9921875,
       "count": "485206"
      },
      {
       "duration": "0.057499647s",
       "percentile": 0.99296875,
       "count": "485587"
      },
      {
       "duration": "0.057905151s",
       "percentile": 0.99375,
       "count": "485970"
      },
      {
       "duration": "0.058208255s",
       "percentile": 0.99453125,
       "count": "486351"
      },
      {
       "duration": "0.058564607s",
       "percentile": 0.9953125,
       "count": "486734"
      },
      {
       "duration": "0.058859519s",
       "percentile": 0.99609375,
       "count": "487115"
      },
      {
       "duration": "0.059033599s",
       "percentile": 0.996484375,
       "count": "487307"
      },
      {
       "duration": "0.059203583s",
       "percentile": 0.996875,
       "count": "487499"
      },
      {
       "duration": "0.059371519s",
       "percentile": 0.997265625,
       "count": "487688"
      },
      {
       "duration": "0.059602943s",
       "percentile": 0.99765625,
       "count": "487880"
      },
      {
       "duration": "0.059772927s",
       "percentile": 0.998046875,
       "count": "488070"
      },
      {
       "duration": "0.059893759s",
       "percentile": 0.9982421875,
       "count": "488168"
      },
      {
       "duration": "0.060012543s",
       "percentile": 0.9984375,
       "count": "488261"
      },
      {
       "duration": "0.060149759s",
       "percentile": 0.9986328125,
       "count": "488357"
      },
      {
       "duration": "0.060289023s",
       "percentile": 0.998828125,
       "count": "488455"
      },
      {
       "duration": "0.060461055s",
       "percentile": 0.9990234375,
       "count": "488548"
      },
      {
       "duration": "0.060547071s",
       "percentile": 0.99912109375,
       "count": "488596"
      },
      {
       "duration": "0.060649471s",
       "percentile": 0.99921875,
       "count": "488644"
      },
      {
       "duration": "0.060764159s",
       "percentile": 0.99931640625,
       "count": "488691"
      },
      {
       "duration": "0.060905471s",
       "percentile": 0.9994140625,
       "count": "488739"
      },
      {
       "duration": "0.061028351s",
       "percentile": 0.99951171875,
       "count": "488787"
      },
      {
       "duration": "0.061083647s",
       "percentile": 0.999560546875,
       "count": "488811"
      },
      {
       "duration": "0.061151231s",
       "percentile": 0.999609375,
       "count": "488834"
      },
      {
       "duration": "0.061229055s",
       "percentile": 0.999658203125,
       "count": "488858"
      },
      {
       "duration": "0.061325311s",
       "percentile": 0.99970703125,
       "count": "488882"
      },
      {
       "duration": "0.061521919s",
       "percentile": 0.999755859375,
       "count": "488906"
      },
      {
       "duration": "0.061638655s",
       "percentile": 0.9997802734375,
       "count": "488918"
      },
      {
       "duration": "0.061755391s",
       "percentile": 0.9998046875,
       "count": "488930"
      },
      {
       "duration": "0.061833215s",
       "percentile": 0.9998291015625,
       "count": "488942"
      },
      {
       "duration": "0.062103551s",
       "percentile": 0.999853515625,
       "count": "488954"
      },
      {
       "duration": "0.062269439s",
       "percentile": 0.9998779296875,
       "count": "488966"
      },
      {
       "duration": "0.062425087s",
       "percentile": 0.99989013671875,
       "count": "488972"
      },
      {
       "duration": "0.062658559s",
       "percentile": 0.99990234375,
       "count": "488978"
      },
      {
       "duration": "0.063094783s",
       "percentile": 0.99991455078125,
       "count": "488984"
      },
      {
       "duration": "0.063416319s",
       "percentile": 0.9999267578125,
       "count": "488990"
      },
      {
       "duration": "0.063911935s",
       "percentile": 0.99993896484375,
       "count": "488996"
      },
      {
       "duration": "0.067158015s",
       "percentile": 0.999945068359375,
       "count": "488999"
      },
      {
       "duration": "0.067620863s",
       "percentile": 0.999951171875,
       "count": "489002"
      },
      {
       "duration": "0.070000639s",
       "percentile": 0.999957275390625,
       "count": "489005"
      },
      {
       "duration": "0.076066815s",
       "percentile": 0.99996337890625,
       "count": "489008"
      },
      {
       "duration": "0.083455999s",
       "percentile": 0.999969482421875,
       "count": "489011"
      },
      {
       "duration": "0.123715583s",
       "percentile": 0.99997253417968746,
       "count": "489012"
      },
      {
       "duration": "0.123949055s",
       "percentile": 0.9999755859375,
       "count": "489014"
      },
      {
       "duration": "0.124588031s",
       "percentile": 0.99997863769531248,
       "count": "489015"
      },
      {
       "duration": "0.473071615s",
       "percentile": 0.999981689453125,
       "count": "489018"
      },
      {
       "duration": "0.473071615s",
       "percentile": 0.9999847412109375,
       "count": "489018"
      },
      {
       "duration": "0.473235455s",
       "percentile": 0.99998626708984373,
       "count": "489019"
      },
      {
       "duration": "0.473268223s",
       "percentile": 0.99998779296875,
       "count": "489020"
      },
      {
       "duration": "0.473268223s",
       "percentile": 0.99998931884765629,
       "count": "489020"
      },
      {
       "duration": "0.473530367s",
       "percentile": 0.99999084472656252,
       "count": "489021"
      },
      {
       "duration": "0.473546751s",
       "percentile": 0.99999237060546875,
       "count": "489022"
      },
      {
       "duration": "0.473546751s",
       "percentile": 0.99999313354492192,
       "count": "489022"
      },
      {
       "duration": "0.473645055s",
       "percentile": 0.999993896484375,
       "count": "489023"
      },
      {
       "duration": "0.473645055s",
       "percentile": 0.99999465942382815,
       "count": "489023"
      },
      {
       "duration": "0.473645055s",
       "percentile": 0.99999542236328121,
       "count": "489023"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.99999618530273438,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.999996566772461,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.99999694824218754,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.999997329711914,
       "count": "489024"
      },
      {
       "duration": "0.473661439s",
       "percentile": 0.9999977111816406,
       "count": "489024"
      },
      {
       "duration": "0.473694207s",
       "percentile": 0.99999809265136719,
       "count": "489025"
      },
      {
       "duration": "0.473694207s",
       "percentile": 1,
       "count": "489025"
      }
     ],
     "min": "0.000135160s",
     "max": "0.473694207s"
    },
    {
     "count": "489025",
     "id": "benchmark_http_client.response_body_size",
     "percentiles": [],
     "raw_mean": 10,
     "raw_pstdev": 0,
     "raw_min": "10",
     "raw_max": "10"
    },
    {
     "count": "489025",
     "id": "benchmark_http_client.response_header_size",
     "percentiles": [],
     "raw_mean": 110,
     "raw_pstdev": 0,
     "raw_min": "110",
     "raw_max": "110"
    },
    {
     "count": "489026",
     "id": "sequencer.blocking",
     "mean": "0.001224014s",
     "pstdev": "0.006619735s",
     "percentiles": [
      {
       "duration": "0.000059835s",
       "percentile": 0,
       "count": "1"
      },
      {
       "duration": "0.000333167s",
       "percentile": 0.1,
       "count": "48905"
      },
      {
       "duration": "0.000365551s",
       "percentile": 0.2,
       "count": "97860"
      },
      {
       "duration": "0.000393375s",
       "percentile": 0.3,
       "count": "146714"
      },
      {
       "duration": "0.000421599s",
       "percentile": 0.4,
       "count": "195634"
      },
      {
       "duration": "0.000452175s",
       "percentile": 0.5,
       "count": "244568"
      },
      {
       "duration": "0.000468927s",
       "percentile": 0.55,
       "count": "269011"
      },
      {
       "duration": "0.000487167s",
       "percentile": 0.6,
       "count": "293443"
      },
      {
       "duration": "0.000507839s",
       "percentile": 0.65,
       "count": "317871"
      },
      {
       "duration": "0.000532031s",
       "percentile": 0.7,
       "count": "342321"
      },
      {
       "duration": "0.000561471s",
       "percentile": 0.75,
       "count": "366794"
      },
      {
       "duration": "0.000579359s",
       "percentile": 0.775,
       "count": "379018"
      },
      {
       "duration": "0.000600543s",
       "percentile": 0.8,
       "count": "391224"
      },
      {
       "duration": "0.000626751s",
       "percentile": 0.825,
       "count": "403461"
      },
      {
       "duration": "0.000659487s",
       "percentile": 0.85,
       "count": "415673"
      },
      {
       "duration": "0.000704639s",
       "percentile": 0.875,
       "count": "427904"
      },
      {
       "duration": "0.000734431s",
       "percentile": 0.8875,
       "count": "434020"
      },
      {
       "duration": "0.000772031s",
       "percentile": 0.9,
       "count": "440130"
      },
      {
       "duration": "0.000821279s",
       "percentile": 0.9125,
       "count": "446239"
      },
      {
       "duration": "0.000885919s",
       "percentile": 0.925,
       "count": "452353"
      },
      {
       "duration": "0.000980767s",
       "percentile": 0.9375,
       "count": "458462"
      },
      {
       "duration": "0.001043487s",
       "percentile": 0.94375,
       "count": "461521"
      },
      {
       "duration": "0.001126783s",
       "percentile": 0.95,
       "count": "464576"
      },
      {
       "duration": "0.001241471s",
       "percentile": 0.95625,
       "count": "467633"
      },
      {
       "duration": "0.001401343s",
       "percentile": 0.9625,
       "count": "470688"
      },
      {
       "duration": "0.001662847s",
       "percentile": 0.96875,
       "count": "473745"
      },
      {
       "duration": "0.001877631s",
       "percentile": 0.971875,
       "count": "475274"
      },
      {
       "duration": "0.002161663s",
       "percentile": 0.975,
       "count": "476801"
      },
      {
       "duration": "0.002608127s",
       "percentile": 0.978125,
       "count": "478329"
      },
      {
       "duration": "0.003245823s",
       "percentile": 0.98125,
       "count": "479857"
      },
      {
       "duration": "0.004198399s",
       "percentile": 0.984375,
       "count": "481385"
      },
      {
       "duration": "0.005203455s",
       "percentile": 0.9859375,
       "count": "482150"
      },
      {
       "duration": "0.008499711s",
       "percentile": 0.9875,
       "count": "482914"
      },
      {
       "duration": "0.049262591s",
       "percentile": 0.9890625,
       "count": "483678"
      },
      {
       "duration": "0.055615487s",
       "percentile": 0.990625,
       "count": "484443"
      },
      {
       "duration": "0.057038847s",
       "percentile": 0.9921875,
       "count": "485206"
      },
      {
       "duration": "0.057520127s",
       "percentile": 0.99296875,
       "count": "485589"
      },
      {
       "duration": "0.057929727s",
       "percentile": 0.99375,
       "count": "485971"
      },
      {
       "duration": "0.058226687s",
       "percentile": 0.99453125,
       "count": "486352"
      },
      {
       "duration": "0.058587135s",
       "percentile": 0.9953125,
       "count": "486735"
      },
      {
       "duration": "0.058877951s",
       "percentile": 0.99609375,
       "count": "487117"
      },
      {
       "duration": "0.059052031s",
       "percentile": 0.996484375,
       "count": "487309"
      },
      {
       "duration": "0.059219967s",
       "percentile": 0.996875,
       "count": "487499"
      },
      {
       "duration": "0.059394047s",
       "percentile": 0.997265625,
       "count": "487692"
      },
      {
       "duration": "0.059623423s",
       "percentile": 0.99765625,
       "count": "487880"
      },
      {
       "duration": "0.059797503s",
       "percentile": 0.998046875,
       "count": "488072"
      },
      {
       "duration": "0.059914239s",
       "percentile": 0.9982421875,
       "count": "488170"
      },
      {
       "duration": "0.060033023s",
       "percentile": 0.9984375,
       "count": "488263"
      },
      {
       "duration": "0.060168191s",
       "percentile": 0.9986328125,
       "count": "488358"
      },
      {
       "duration": "0.060305407s",
       "percentile": 0.998828125,
       "count": "488453"
      },
      {
       "duration": "0.060481535s",
       "percentile": 0.9990234375,
       "count": "488549"
      },
      {
       "duration": "0.060565503s",
       "percentile": 0.99912109375,
       "count": "488597"
      },
      {
       "duration": "0.060671999s",
       "percentile": 0.99921875,
       "count": "488644"
      },
      {
       "duration": "0.060786687s",
       "percentile": 0.99931640625,
       "count": "488692"
      },
      {
       "duration": "0.060925951s",
       "percentile": 0.9994140625,
       "count": "488740"
      },
      {
       "duration": "0.061059071s",
       "percentile": 0.99951171875,
       "count": "488788"
      },
      {
       "duration": "0.061114367s",
       "percentile": 0.999560546875,
       "count": "488812"
      },
      {
       "duration": "0.061171711s",
       "percentile": 0.999609375,
       "count": "488835"
      },
      {
       "duration": "0.061255679s",
       "percentile": 0.999658203125,
       "count": "488859"
      },
      {
       "duration": "0.061353983s",
       "percentile": 0.99970703125,
       "count": "488883"
      },
      {
       "duration": "0.061550591s",
       "percentile": 0.999755859375,
       "count": "488907"
      },
      {
       "duration": "0.061663231s",
       "percentile": 0.9997802734375,
       "count": "488919"
      },
      {
       "duration": "0.061771775s",
       "percentile": 0.9998046875,
       "count": "488931"
      },
      {
       "duration": "0.061878271s",
       "percentile": 0.9998291015625,
       "count": "488943"
      },
      {
       "duration": "0.062142463s",
       "percentile": 0.999853515625,
       "count": "488955"
      },
      {
       "duration": "0.062314495s",
       "percentile": 0.9998779296875,
       "count": "488967"
      },
      {
       "duration": "0.062445567s",
       "percentile": 0.99989013671875,
       "count": "488973"
      },
      {
       "duration": "0.062691327s",
       "percentile": 0.99990234375,
       "count": "488979"
      },
      {
       "duration": "0.063129599s",
       "percentile": 0.99991455078125,
       "count": "488985"
      },
      {
       "duration": "0.063440895s",
       "percentile": 0.9999267578125,
       "count": "488991"
      },
      {
       "duration": "0.064053247s",
       "percentile": 0.99993896484375,
       "count": "488997"
      },
      {
       "duration": "0.067235839s",
       "percentile": 0.999945068359375,
       "count": "489000"
      },
      {
       "duration": "0.067674111s",
       "percentile": 0.999951171875,
       "count": "489003"
      },
      {
       "duration": "0.070094847s",
       "percentile": 0.999957275390625,
       "count": "489006"
      },
      {
       "duration": "0.076132351s",
       "percentile": 0.99996337890625,
       "count": "489009"
      },
      {
       "duration": "0.083533823s",
       "percentile": 0.999969482421875,
       "count": "489012"
      },
      {
       "duration": "0.123760639s",
       "percentile": 0.99997253417968746,
       "count": "489013"
      },
      {
       "duration": "0.124637183s",
       "percentile": 0.9999755859375,
       "count": "489015"
      },
      {
       "duration": "0.124870655s",
       "percentile": 0.99997863769531248,
       "count": "489016"
      },
      {
       "duration": "0.564232191s",
       "percentile": 0.999981689453125,
       "count": "489018"
      },
      {
       "duration": "0.564363263s",
       "percentile": 0.9999847412109375,
       "count": "489019"
      },
      {
       "duration": "0.564527103s",
       "percentile": 0.99998626708984373,
       "count": "489020"
      },
      {
       "duration": "0.564658175s",
       "percentile": 0.99998779296875,
       "count": "489021"
      },
      {
       "duration": "0.564658175s",
       "percentile": 0.99998931884765629,
       "count": "489021"
      },
      {
       "duration": "0.564690943s",
       "percentile": 0.99999084472656252,
       "count": "489022"
      },
      {
       "duration": "0.564920319s",
       "percentile": 0.99999237060546875,
       "count": "489023"
      },
      {
       "duration": "0.564920319s",
       "percentile": 0.99999313354492192,
       "count": "489023"
      },
      {
       "duration": "0.564985855s",
       "percentile": 0.999993896484375,
       "count": "489024"
      },
      {
       "duration": "0.564985855s",
       "percentile": 0.99999465942382815,
       "count": "489024"
      },
      {
       "duration": "0.564985855s",
       "percentile": 0.99999542236328121,
       "count": "489024"
      },
      {
       "duration": "0.565051391s",
       "percentile": 0.99999618530273438,
       "count": "489025"
      },
      {
       "duration": "0.565051391s",
       "percentile": 0.999996566772461,
       "count": "489025"
      },
      {
       "duration": "0.565051391s",
       "percentile": 0.99999694824218754,
       "count": "489025"
      },
      {
       "duration": "0.565051391s",
       "percentile": 0.999997329711914,
       "count": "489025"
      },
      {
       "duration": "0.565051391s",
       "percentile": 0.9999977111816406,
       "count": "489025"
      },
      {
       "duration": "0.565084159s",
       "percentile": 0.99999809265136719,
       "count": "489026"
      },
      {
       "duration": "0.565084159s",
       "percentile": 1,
       "count": "489026"
      }
     ],
     "min": "0.000059834s",
     "max": "0.565084159s"
    },
    {
     "count": "490015",
     "id": "sequencer.callback",
     "mean": "0.001504828s",
     "pstdev": "0.009302005s",
     "percentiles": [
      {
       "duration": "0.000143007s",
       "percentile": 0,
       "count": "1"
      },
      {
       "duration": "0.000329759s",
       "percentile": 0.1,
       "count": "49052"
      },
      {
       "duration": "0.000362047s",
       "percentile": 0.2,
       "count": "98067"
      },
      {
       "duration": "0.000389887s",
       "percentile": 0.3,
       "count": "147024"
      },
      {
       "duration": "0.000418095s",
       "percentile": 0.4,
       "count": "196013"
      },
      {
       "duration": "0.000448639s",
       "percentile": 0.5,
       "count": "245054"
      },
      {
       "duration": "0.000465423s",
       "percentile": 0.55,
       "count": "269562"
      },
      {
       "duration": "0.000483711s",
       "percentile": 0.6,
       "count": "294018"
      },
      {
       "duration": "0.000504431s",
       "percentile": 0.65,
       "count": "318511"
      },
      {
       "duration": "0.000528735s",
       "percentile": 0.7,
       "count": "343028"
      },
      {
       "duration": "0.000558175s",
       "percentile": 0.75,
       "count": "367536"
      },
      {
       "duration": "0.000576191s",
       "percentile": 0.775,
       "count": "379765"
      },
      {
       "duration": "0.000597631s",
       "percentile": 0.8,
       "count": "392012"
      },
      {
       "duration": "0.000624095s",
       "percentile": 0.825,
       "count": "404284"
      },
      {
       "duration": "0.000657311s",
       "percentile": 0.85,
       "count": "416522"
      },
      {
       "duration": "0.000703391s",
       "percentile": 0.875,
       "count": "428769"
      },
      {
       "duration": "0.000733855s",
       "percentile": 0.8875,
       "count": "434889"
      },
      {
       "duration": "0.000772927s",
       "percentile": 0.9,
       "count": "441021"
      },
      {
       "duration": "0.000823807s",
       "percentile": 0.9125,
       "count": "447140"
      },
      {
       "duration": "0.000892063s",
       "percentile": 0.925,
       "count": "453264"
      },
      {
       "duration": "0.000991519s",
       "percentile": 0.9375,
       "count": "459393"
      },
      {
       "duration": "0.001059199s",
       "percentile": 0.94375,
       "count": "462453"
      },
      {
       "duration": "0.001149055s",
       "percentile": 0.95,
       "count": "465515"
      },
      {
       "duration": "0.001276415s",
       "percentile": 0.95625,
       "count": "468577"
      },
      {
       "duration": "0.001459263s",
       "percentile": 0.9625,
       "count": "471640"
      },
      {
       "duration": "0.001775103s",
       "percentile": 0.96875,
       "count": "474703"
      },
      {
       "duration": "0.002026367s",
       "percentile": 0.971875,
       "count": "476234"
      },
      {
       "duration": "0.002416127s",
       "percentile": 0.975,
       "count": "477765"
      },
      {
       "duration": "0.002970495s",
       "percentile": 0.978125,
       "count": "479296"
      },
      {
       "duration": "0.003781119s",
       "percentile": 0.98125,
       "count": "480828"
      },
      {
       "duration": "0.005635327s",
       "percentile": 0.984375,
       "count": "482359"
      },
      {
       "duration": "0.013006847s",
       "percentile": 0.9859375,
       "count": "483125"
      },
      {
       "duration": "0.052420607s",
       "percentile": 0.9875,
       "count": "483890"
      },
      {
       "duration": "0.056158207s",
       "percentile": 0.9890625,
       "count": "484656"
      },
      {
       "duration": "0.057298943s",
       "percentile": 0.990625,
       "count": "485424"
      },
      {
       "duration": "0.058091519s",
       "percentile": 0.9921875,
       "count": "486190"
      },
      {
       "duration": "0.058400767s",
       "percentile": 0.99296875,
       "count": "486571"
      },
      {
       "duration": "0.058763263s",
       "percentile": 0.99375,
       "count": "486956"
      },
      {
       "duration": "0.059068415s",
       "percentile": 0.99453125,
       "count": "487337"
      },
      {
       "duration": "0.059420671s",
       "percentile": 0.9953125,
       "count": "487720"
      },
      {
       "duration": "0.059815935s",
       "percentile": 0.99609375,
       "count": "488103"
      },
      {
       "duration": "0.060073983s",
       "percentile": 0.996484375,
       "count": "488294"
      },
      {
       "duration": "0.060354559s",
       "percentile": 0.996875,
       "count": "488485"
      },
      {
       "duration": "0.060747775s",
       "percentile": 0.997265625,
       "count": "488677"
      },
      {
       "duration": "0.061278207s",
       "percentile": 0.99765625,
       "count": "488867"
      },
      {
       "duration": "0.140812287s",
       "percentile": 0.998046875,
       "count": "489059"
      },
      {
       "duration": "0.140918783s",
       "percentile": 0.9982421875,
       "count": "489154"
      },
      {
       "duration": "0.141049855s",
       "percentile": 0.9984375,
       "count": "489254"
      },
      {
       "duration": "0.141099007s",
       "percentile": 0.9986328125,
       "count": "489352"
      },
      {
       "duration": "0.141148159s",
       "percentile": 0.998828125,
       "count": "489452"
      },
      {
       "duration": "0.141213695s",
       "percentile": 0.9990234375,
       "count": "489541"
      },
      {
       "duration": "0.141246463s",
       "percentile": 0.99912109375,
       "count": "489588"
      },
      {
       "duration": "0.141262847s",
       "percentile": 0.99921875,
       "count": "489639"
      },
      {
       "duration": "0.141287423s",
       "percentile": 0.99931640625,
       "count": "489696"
      },
      {
       "duration": "0.141311999s",
       "percentile": 0.9994140625,
       "count": "489739"
      },
      {
       "duration": "0.141336575s",
       "percentile": 0.99951171875,
       "count": "489792"
      },
      {
       "duration": "0.141344767s",
       "percentile": 0.999560546875,
       "count": "489809"
      },
      {
       "duration": "0.141352959s",
       "percentile": 0.999609375,
       "count": "489824"
      },
      {
       "duration": "0.141369343s",
       "percentile": 0.999658203125,
       "count": "489852"
      },
      {
       "duration": "0.141377535s",
       "percentile": 0.99970703125,
       "count": "489873"
      },
      {
       "duration": "0.141402111s",
       "percentile": 0.999755859375,
       "count": "489900"
      },
      {
       "duration": "0.141418495s",
       "percentile": 0.9997802734375,
       "count": "489908"
      },
      {
       "duration": "0.141443071s",
       "percentile": 0.9998046875,
       "count": "489922"
      },
      {
       "duration": "0.141459455s",
       "percentile": 0.9998291015625,
       "count": "489933"
      },
      {
       "duration": "0.141475839s",
       "percentile": 0.999853515625,
       "count": "489944"
      },
      {
       "duration": "0.141492223s",
       "percentile": 0.9998779296875,
       "count": "489956"
      },
      {
       "duration": "0.141647871s",
       "percentile": 0.99989013671875,
       "count": "489965"
      },
      {
       "duration": "0.141656063s",
       "percentile": 0.99990234375,
       "count": "489969"
      },
      {
       "duration": "0.141664255s",
       "percentile": 0.99991455078125,
       "count": "489981"
      },
      {
       "duration": "0.141664255s",
       "percentile": 0.9999267578125,
       "count": "489981"
      },
      {
       "duration": "0.141672447s",
       "percentile": 0.99993896484375,
       "count": "489987"
      },
      {
       "duration": "0.141680639s",
       "percentile": 0.999945068359375,
       "count": "489996"
      },
      {
       "duration": "0.141680639s",
       "percentile": 0.999951171875,
       "count": "489996"
      },
      {
       "duration": "0.141680639s",
       "percentile": 0.999957275390625,
       "count": "489996"
      },
      {
       "duration": "0.141688831s",
       "percentile": 0.99996337890625,
       "count": "490005"
      },
      {
       "duration": "0.141688831s",
       "percentile": 0.999969482421875,
       "count": "490005"
      },
      {
       "duration": "0.141688831s",
       "percentile": 0.99997253417968746,
       "count": "490005"
      },
      {
       "duration": "0.141688831s",
       "percentile": 0.9999755859375,
       "count": "490005"
      },
      {
       "duration": "0.141688831s",
       "percentile": 0.99997863769531248,
       "count": "490005"
      },
      {
       "duration": "0.694484991s",
       "percentile": 0.999981689453125,
       "count": "490010"
      },
      {
       "duration": "0.694484991s",
       "percentile": 0.9999847412109375,
       "count": "490010"
      },
      {
       "duration": "0.694484991s",
       "percentile": 0.99998626708984373,
       "count": "490010"
      },
      {
       "duration": "0.694484991s",
       "percentile": 0.99998779296875,
       "count": "490010"
      },
      {
       "duration": "0.694484991s",
       "percentile": 0.99998931884765629,
       "count": "490010"
      },
      {
       "duration": "0.694550527s",
       "percentile": 0.99999084472656252,
       "count": "490011"
      },
      {
       "duration": "0.694648831s",
       "percentile": 0.99999237060546875,
       "count": "490012"
      },
      {
       "duration": "0.694648831s",
       "percentile": 0.99999313354492192,
       "count": "490012"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.999993896484375,
       "count": "490014"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.99999465942382815,
       "count": "490014"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.99999542236328121,
       "count": "490014"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.99999618530273438,
       "count": "490014"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.999996566772461,
       "count": "490014"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.99999694824218754,
       "count": "490014"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.999997329711914,
       "count": "490014"
      },
      {
       "duration": "0.694714367s",
       "percentile": 0.9999977111816406,
       "count": "490014"
      },
      {
       "duration": "0.694747135s",
       "percentile": 0.99999809265136719,
       "count": "490015"
      },
      {
       "duration": "0.694747135s",
       "percentile": 1,
       "count": "490015"
      }
     ],
     "min": "0.000143s",
     "max": "0.694747135s"
    }
   ],
   "counters": [
    {
     "name": "benchmark.http_2xx",
     "value": "489025"
    },
    {
     "name": "benchmark.pool_overflow",
     "value": "990"
    },
    {
     "name": "cluster_manager.cluster_added",
     "value": "10"
    },
    {
     "name": "default.total_match_count",
     "value": "10"
    },
    {
     "name": "membership_change",
     "value": "10"
    },
    {
     "name": "runtime.load_success",
     "value": "1"
    },
    {
     "name": "runtime.override_dir_not_exists",
     "value": "1"
    },
    {
     "name": "upstream_cx_http1_total",
     "value": "10"
    },
    {
     "name": "upstream_cx_rx_bytes_total",
     "value": "76776925"
    },
    {
     "name": "upstream_cx_total",
     "value": "10"
    },
    {
     "name": "upstream_cx_tx_bytes_total",
     "value": "22006170"
    },
    {
     "name": "upstream_rq_pending_overflow",
     "value": "990"
    },
    {
     "name": "upstream_rq_pending_total",
     "value": "10"
    },
    {
     "name": "upstream_rq_total",
     "value": "489026"
    }
   ],
   "execution_duration": "60.000239322s",
   "execution_start": "2025-08-19T03:58:43.661836460Z",
   "user_defined_outputs": []
  }`)

	var r proto.Result
	if err := protojson.Unmarshal(data, &r); err != nil {
		return &proto.Result{}
	}

	return &r
}

func TestToMarkdown(t *testing.T) {
	input := &BenchmarkSuiteReport{
		Settings: map[string]string{
			"rps":        "1000",
			"connection": "100",
			"duration":   "30",
			"cpu":        "1000m",
			"memory":     "1000Mi",
		},
		Reports: []*BenchmarkTestReport{
			{
				Title:       "fake-title",
				Description: "fake-description",
				Reports: []*BenchmarkCaseReport{
					{
						Title:             "case-title",
						Routes:            100,
						RoutesPerHostname: 2,
						Result:            fakeCaseResult(),
						Phase:             "fake-phase",
						HeapProfileImage:  "fake-image",
					},
				},
			},
		},
	}
	out, err := ToMarkdown(input)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	if test.OverrideTestData() {
		_ = os.WriteFile("testdata/markdown_output.md", out, 0600)
		return
	}

	data, err := os.ReadFile("testdata/markdown_output.md")
	require.NoError(t, err)
	require.Equal(t, string(data), string(out))
}
