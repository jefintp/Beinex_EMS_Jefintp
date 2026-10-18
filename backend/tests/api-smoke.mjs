// End-to-end API test against the RUNNING backend and real database.
//
//   1. Start the backend (go run ./cmd/server)
//   2. In another terminal, from the backend folder:  node tests/api-smoke.mjs
//
// Needs Node 18+. It creates its own test records (unique emails) and deletes
// them at the end, so it is safe to run against your dev database.

const BASE = process.env.API_URL || 'http://localhost:8080/api/v1';
const unique = Date.now();
const email = (tag) => `smoke.${tag}.${unique}@example.com`;

const body = (over = {}) => ({
  name: 'Smoke Test',
  email: email('a'),
  department: 'QA-Smoke',
  designation: 'Tester',
  joining_date: '2024-06-15',
  status: 'active',
  ...over,
});

async function call(method, path, payload) {
  const res = await fetch(BASE + path, {
    method,
    headers: payload ? { 'Content-Type': 'application/json' } : {},
    body: payload ? JSON.stringify(payload) : undefined,
  });
  const text = await res.text();
  let json = null;
  try { json = text ? JSON.parse(text) : null; } catch { /* not JSON */ }
  return { status: res.status, json, text };
}

let passed = 0;
let failed = 0;
const createdIds = [];

async function test(name, fn) {
  try {
    await fn();
    passed++;
    console.log(`  PASS  ${name}`);
  } catch (e) {
    failed++;
    console.log(`  FAIL  ${name}\n        ${e.message}`);
  }
}

function eq(actual, expected, label = 'value') {
  if (actual !== expected) throw new Error(`${label}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`);
}
function ok(cond, msg) { if (!cond) throw new Error(msg); }

async function main() {
  console.log(`Testing ${BASE}\n`);

  try {
    const h = await fetch(BASE.replace(/\/api\/v1$/, '') + '/health');
    eq(h.status, 200, 'health status');
  } catch {
    console.log('Cannot reach the backend. Is `go run ./cmd/server` running?');
    process.exit(2);
  }

  let id;

  console.log('Create');
  await test('POST valid employee -> 201 with id', async () => {
    const r = await call('POST', '/employees', body());
    eq(r.status, 201, 'status');
    ok(r.json && r.json.id > 0, 'response has id');
    eq(r.json.status, 'active', 'status field');
    eq(r.json.joining_date, '2024-06-15', 'joining_date');
    id = r.json.id;
    createdIds.push(id);
  });
  await test('POST without status -> defaults to active', async () => {
    const payload = body({ email: email('nostatus') });
    delete payload.status;
    const r = await call('POST', '/employees', payload);
    eq(r.status, 201, 'status');
    eq(r.json.status, 'active', 'default status');
    createdIds.push(r.json.id);
  });
  await test('POST email is stored lowercase', async () => {
    const r = await call('POST', '/employees', body({ email: email('UPPER').toUpperCase() }));
    eq(r.status, 201, 'status');
    eq(r.json.email, r.json.email.toLowerCase(), 'email case');
    createdIds.push(r.json.id);
  });

  console.log('\nValidation');
  await test('empty body -> 400 with field errors', async () => {
    const r = await call('POST', '/employees', {});
    eq(r.status, 400, 'status');
    ok(r.json.fields && r.json.fields.name && r.json.fields.email, 'fields present');
  });
  for (const bad of ['abc', 'a@b', 'a@b.', 'Bob <bob@example.com>']) {
    await test(`invalid email "${bad}" -> 400`, async () => {
      const r = await call('POST', '/employees', body({ email: bad }));
      eq(r.status, 400, 'status');
      ok(r.json.fields && r.json.fields.email, 'email field error');
    });
  }
  await test('invalid date -> 400', async () => {
    const r = await call('POST', '/employees', body({ email: email('d'), joining_date: '2024-13-45' }));
    eq(r.status, 400, 'status');
    ok(r.json.fields.joining_date, 'joining_date error');
  });
  await test('future joining date -> 400', async () => {
    const r = await call('POST', '/employees', body({ email: email('f'), joining_date: '2999-01-01' }));
    eq(r.status, 400, 'status');
    ok(r.json.fields.joining_date, 'joining_date error');
  });
  await test('invalid status -> 400', async () => {
    const r = await call('POST', '/employees', body({ email: email('s'), status: 'pending' }));
    eq(r.status, 400, 'status');
    ok(r.json.fields.status, 'status error');
  });
  await test('malformed JSON -> 400', async () => {
    const res = await fetch(BASE + '/employees', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{oops' });
    eq(res.status, 400, 'status');
  });

  console.log('\nDuplicate email');
  await test('same email again -> 409', async () => {
    const r = await call('POST', '/employees', body());
    eq(r.status, 409, 'status');
    ok(r.json.fields && r.json.fields.email, 'email field message');
  });
  await test('same email in different case -> 409', async () => {
    const r = await call('POST', '/employees', body({ email: email('a').toUpperCase() }));
    eq(r.status, 409, 'status');
  });

  console.log('\nRead');
  await test('GET by id -> 200', async () => {
    const r = await call('GET', `/employees/${id}`);
    eq(r.status, 200, 'status');
    eq(r.json.id, id, 'id');
  });
  await test('GET unknown id -> 404', async () => eq((await call('GET', '/employees/99999999')).status, 404, 'status'));
  await test('GET non-numeric id -> 400', async () => eq((await call('GET', '/employees/abc')).status, 400, 'status'));
  await test('GET list -> 200 with {data,total}', async () => {
    const r = await call('GET', '/employees');
    eq(r.status, 200, 'status');
    ok(Array.isArray(r.json.data), 'data is an array');
    ok(r.json.total >= 3, 'total counts created rows');
  });

  console.log('\nSearch and filters');
  await test('search by email fragment finds the record', async () => {
    const r = await call('GET', `/employees?search=smoke.a.${unique}`);
    eq(r.json.data.length, 1, 'result count');
  });
  await test('search is case-insensitive', async () => {
    const r = await call('GET', `/employees?search=SMOKE.A.${unique}`);
    eq(r.json.data.length, 1, 'result count');
  });
  await test('search by exact id finds the record', async () => {
    const r = await call('GET', `/employees?search=${id}`);
    eq(r.status, 200, 'status');
    ok(r.json.data.some((e) => e.id === id), `employee ${id} is in the results`);
  });
  await test('search with no match -> empty list (not null)', async () => {
    const r = await call('GET', '/employees?search=zzz-no-such-person-zzz');
    eq(r.status, 200, 'status');
    eq(r.json.data.length, 0, 'result count');
  });
  await test('filter by department', async () => {
    const r = await call('GET', '/employees?department=QA-Smoke');
    ok(r.json.data.length >= 3, 'at least our 3 records');
    ok(r.json.data.every((e) => e.department === 'QA-Smoke'), 'all match department');
  });
  await test('filter by status=inactive excludes active rows', async () => {
    const r = await call('GET', '/employees?department=QA-Smoke&status=inactive');
    eq(r.json.data.length, 0, 'result count');
  });
  await test('departments endpoint lists QA-Smoke', async () => {
    const r = await call('GET', '/departments');
    eq(r.status, 200, 'status');
    ok(r.json.data.includes('QA-Smoke'), 'department present');
  });

  console.log('\nUpdate');
  await test('PUT valid change -> 200 and persisted', async () => {
    const r = await call('PUT', `/employees/${id}`, body({ designation: 'Senior Tester', status: 'inactive' }));
    eq(r.status, 200, 'status');
    eq(r.json.designation, 'Senior Tester', 'designation');
    const again = await call('GET', `/employees/${id}`);
    eq(again.json.status, 'inactive', 'persisted status');
  });
  await test('filter by status=inactive now finds it', async () => {
    const r = await call('GET', '/employees?department=QA-Smoke&status=inactive');
    eq(r.json.data.length, 1, 'result count');
  });
  await test('PUT with another employee\'s email -> 409', async () => {
    const r = await call('PUT', `/employees/${id}`, body({ email: email('nostatus') }));
    eq(r.status, 409, 'status');
  });
  await test('PUT unknown id -> 404', async () => eq((await call('PUT', '/employees/99999999', body({ email: email('x') }))).status, 404, 'status'));
  await test('PUT invalid data -> 400', async () => eq((await call('PUT', `/employees/${id}`, body({ email: 'nope' }))).status, 400, 'status'));

  console.log('\nDelete');
  await test('DELETE -> 204', async () => {
    const r = await call('DELETE', `/employees/${id}`);
    eq(r.status, 204, 'status');
    createdIds.splice(createdIds.indexOf(id), 1);
  });
  await test('GET deleted id -> 404', async () => eq((await call('GET', `/employees/${id}`)).status, 404, 'status'));
  await test('DELETE again -> 404', async () => eq((await call('DELETE', `/employees/${id}`)).status, 404, 'status'));
  await test('deleted email can be reused -> 201', async () => {
    const r = await call('POST', '/employees', body());
    eq(r.status, 201, 'status');
    createdIds.push(r.json.id);
  });
}

main()
  .catch((e) => { console.log('Unexpected error:', e); failed++; })
  .finally(async () => {
    for (const cid of createdIds) await call('DELETE', `/employees/${cid}`).catch(() => {});
    console.log(`\n${passed} passed, ${failed} failed`);
    process.exit(failed ? 1 : 0);
  });
