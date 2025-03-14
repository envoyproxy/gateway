# Envoy Gateway Fuzzing

Envoy Gateway fuzzers are run continuously on [OSS-Fuzz](https://google.github.io/oss-fuzz/). Design documents is
available [here](https://gateway.envoyproxy.io/contributions/design/fuzzing/).

## Local testing

To run the fuzzers locally, you can use the following command:

```bash
git clone --depth=1 https://github.com/google/oss-fuzz.git

cd oss-fuzz

python3 infra/helper.py build_image gateway

python3 infra/helper.py build_fuzzers gateway 
```

To run the fuzzers use `python3 infra/helper.py run_fuzzer gateway <Fuzzer Name>`. For example:

```bash
python3 infra/helper.py run_fuzzer gateway FuzzGatewayAPIToXDS
```