# james-load-testing

`docker run -it -v ./.env:/.env <image> loadtest --run_for_minutes=600 --req_per_second=5 --member_per_group=20`

`helm install james-load-testing ./deploy/jamesloadtest`