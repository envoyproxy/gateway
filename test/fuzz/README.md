# Envoy Gateway Fuzzing

Envoy Gateway fuzzers are run continuously on [OSS-Fuzz](https://google.github.io/oss-fuzz/). Design documents is
available [here](https://gateway.envoyproxy.io/contributions/design/fuzzing/).

## Local testing

### Run the fuzzers locally

Run the following command from the root of the repository.

```bash
make go.test.fuzz FUZZ_TIME=10s
```


### Run the fuzzers locally with OSS-Fuzz infra

To run the fuzzers locally using oss-fuzz infra, you can use the following commands:

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