const axios = require("axios");

const baseURL = "http://localhost:8080";


describe("Rate Limiter tests", () => {
	jest.setTimeout(10000);

	test("GET /foo 5 times should return 200, 6th time it will return 429",
		async () => {

			const client = axios.create({
				headers: {
					Authorization: `Bearer 2`,
				},
				validateStatus: function (status) {
					return status < 500;
				},
			});

			// User 2 can make 5 requests until the bucket empties
			for (let i = 0; i < 5; i++) {
				const res = await client.get(`${baseURL}/foo`);
				expect(res.status).toBe(200);
				expect(res.data).toBeTruthy();
				expect(res.data).toEqual("{success: true}");

			}

			const res = await client.get(`${baseURL}/foo`);
			expect(res.status).toBe(429);
			expect(res.data).toBeTruthy();
			expect(res.data).toContain("rate limit exceeded");
			// sleep 5 seconds to refill the bucket
			await new Promise(r => setTimeout(r, 5000));

		},
	);

	test("GET /foo 5 times should return 200, 6th time it will return 429. Wait a second, make another request. It should return 200",
		async () => {
			const client = axios.create({
				headers: {
					Authorization: `Bearer 2`,
				},
				validateStatus: function (status) {
					return status < 500;
				},
			});
			// User 2 can make 5 requests until the bucket empties. Then, the bucket refills with a rate of 1 token per second for user 2
			for (let i = 0; i < 5; i++) {
				const res = await client.get(`${baseURL}/foo`);
				expect(res.status).toBe(200);
				expect(res.data).toBeTruthy();
				expect(res.data).toEqual("{success: true}");

			}

			let res = await client.get(`${baseURL}/foo`);
			expect(res.status).toBe(429);
			expect(res.data).toBeTruthy();
			expect(res.data).toContain("rate limit exceeded");

			await new Promise(r => setTimeout(r, 1000));

			res = await client.get(`${baseURL}/foo`);
			expect(res.status).toBe(200);
			expect(res.data).toBeTruthy();
			expect(res.data).toEqual("{success: true}");
			await new Promise(r => setTimeout(r, 5000));

		}
	);

	test("GET /foo 5 times with user 2, then 3 times with user 1 to check limits are done per user",
		async () => {
			const client2 = axios.create({
				headers: {
					Authorization: `Bearer 2`,
				},
				validateStatus: function (status) {
					return status < 500;
				},
			});
			// User 2 can make 5 requests until the bucket empties. Then, the bucket refills with a rate of 1 token per second for user 2
			for (let i = 0; i < 5; i++) {
				const res = await client2.get(`${baseURL}/foo`);
				expect(res.status).toBe(200);
				expect(res.data).toBeTruthy();
				expect(res.data).toEqual("{success: true}");

			}

			const client1 = axios.create({
				headers: {
					Authorization: `Bearer 1`,
				},
				validateStatus: function (status) {
					return status < 500;
				},
			});

			for (let i = 0; i < 2; i++) {
				const res = await client1.get(`${baseURL}/foo`);
				expect(res.status).toBe(200);
				expect(res.data).toBeTruthy();
				expect(res.data).toEqual("{success: true}");

			}

			await new Promise(r => setTimeout(r, 5000));

		}
	);

	test("GET /foo 5 times with user 1, wait two seconds, see user 1 can make a single more request",
		async () => {
			const client1 = axios.create({
				headers: {
					Authorization: `Bearer 1`,
				},
				validateStatus: function (status) {
					return status < 500;
				},
			});
			// User 1 can make 5 requests until the bucket empties. Then, the bucket refills with a rate of 0.5 token per second for user 1
			for (let i = 0; i < 5; i++) {
				const res = await client1.get(`${baseURL}/foo`);
				expect(res.status).toBe(200);
				expect(res.data).toBeTruthy();
				expect(res.data).toEqual("{success: true}");

			}

			await new Promise(r => setTimeout(r, 2000));

			let res = await client1.get(`${baseURL}/foo`);
			expect(res.status).toBe(200);

			res = await client1.get(`${baseURL}/foo`);
			expect(res.status).toBe(429);


		}
	);

	
});
