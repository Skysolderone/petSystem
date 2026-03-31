.PHONY: build run test smoke tidy migrate-up migrate-down mobile-install mobile-typecheck mobile-start check

build:
	$(MAKE) -C server build

run:
	$(MAKE) -C server run

test:
	$(MAKE) -C server test

smoke:
	$(MAKE) -C server smoke

tidy:
	$(MAKE) -C server tidy

migrate-up:
	$(MAKE) -C server migrate-up

migrate-down:
	$(MAKE) -C server migrate-down

mobile-install:
	cd apps/mobile && npm install

mobile-typecheck:
	cd apps/mobile && npx tsc --noEmit

mobile-start:
	cd apps/mobile && npx expo start

check: test mobile-typecheck
