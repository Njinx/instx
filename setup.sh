#!/bin/bash

echo "[[ Python dependencies ]]"
/usr/bin/env python -m venv env
source env/bin/activate
/usr/bin/env python -m pip install -r requirements.txt
env/bin/deactivate

echo "[[ Golang dependencies ]]"
pushd "./proxy" || echo "Could not enter \"proxy/\""; exit

go build

popd || echo "Could not exit \"proxy/\""; exit
