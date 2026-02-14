# AGENTS.md

## Stack

TypeScript (strict mode), NestJS, Jest, class-validator, class-transformer.

## Commands

| Command | Purpose |
|---|---|
| `npm run build` | Compile — must pass with zero errors |
| `npm run lint` | ESLint — fix all warnings before committing |
| `npm run format` | Prettier — run before committing |
| `npm run test` | Unit tests |
| `npm run test:e2e` | Integration tests |

Run `build`, `lint`, and `test` before every commit. All must pass.

## Project Structure

```
src/
  common/          # Shared guards, pipes, filters, decorators, interceptors
  config/          # Configuration & env validation
  <feature>/
    <feature>.module.ts
    <feature>.controller.ts
    <feature>.controller.spec.ts
    <feature>.service.ts
    <feature>.service.spec.ts
    <feature>.repository.ts
    dto/
    entities/
test/              # e2e tests (*.e2e-spec.ts)
```

Group by feature, not by type. Shared code goes in `common/`.

## TypeScript Rules

- No `any`. Use `unknown` + type guards.
- No non-null assertions (`!`) without a justifying comment.
- Explicit return types on all public methods.
- Use `readonly` on injected dependencies and immutable properties.
- Use `import type` for type-only imports.
- File names: `kebab-case`. Classes: `PascalCase`. Functions/variables: `camelCase`. Constants: `UPPER_SNAKE_CASE`.

## NestJS Rules

- Constructor injection only. All deps `private readonly`.
- Controllers are thin: validate input → delegate to service → return response.
- Services are stateless. All business logic lives here.
- All request bodies and query params use DTOs with `class-validator` decorators.
- Use `PartialType()` / `PickType()` / `OmitType()` to derive DTOs — don't duplicate fields.
- Throw NestJS built-in exceptions (`NotFoundException`, `BadRequestException`, etc.).
- Never expose stack traces or internal errors to clients.

## Testing — MANDATORY

Every change must include tests. No exceptions.

### Unit Tests

- Co-located: `<name>.spec.ts` next to the source file.
- **Add tests to the existing spec file for that class.** Do not create a new spec file if one exists.
- Mock all external dependencies. Never mock the class under test.
- Use Arrange-Act-Assert pattern.
- Test happy path, error paths, and edge cases.
- Call `jest.clearAllMocks()` in `beforeEach`.

```typescript
describe('OrderService', () => {
  let service: OrderService;
  let repo: jest.Mocked<OrderRepository>;

  beforeEach(async () => {
    const module = await Test.createTestingModule({
      providers: [
        OrderService,
        { provide: OrderRepository, useValue: { save: jest.fn(), findOne: jest.fn() } },
      ],
    }).compile();
    service = module.get(OrderService);
    repo = module.get(OrderRepository);
    jest.clearAllMocks();
  });

  it('should create an order', async () => {
    repo.save.mockResolvedValue(mockOrder);
    const result = await service.create(dto);
    expect(result).toEqual(mockOrder);
    expect(repo.save).toHaveBeenCalledWith(expect.objectContaining({ productId: dto.productId }));
  });

  it('should throw NotFoundException when order not found', async () => {
    repo.findOne.mockResolvedValue(null);
    await expect(service.findOne('missing-id')).rejects.toThrow(NotFoundException);
  });
});
```

### Integration (e2e) Tests

- Located in `test/`, named `*.e2e-spec.ts`.
- Test full HTTP lifecycle with `supertest`.
- Validate status codes, response shapes, and validation rejection.

```typescript
it('POST /users — 201', () => {
  return request(app.getHttpServer())
    .post('/users')
    .send({ name: 'Jane', email: 'jane@example.com' })
    .expect(201)
    .expect((res) => expect(res.body).toHaveProperty('id'));
});

it('POST /users — 400 on invalid body', () => {
  return request(app.getHttpServer()).post('/users').send({}).expect(400);
});
```

### What to Test Per Layer

| Layer | Verify |
|---|---|
| Controller | Status codes, response shape, validation rejection, guard behavior |
| Service | Business logic, branches, exceptions thrown, side effects |
| Repository | Query correctness (e2e with DB) |
| DTOs | Valid data passes, invalid data fails |
| Guards/Pipes | Access control, transformation |

## API Conventions

- RESTful: `GET` read, `POST` create, `PATCH` partial update, `DELETE` remove.
- Plural resource names: `/users`, `/orders`.
- Paginate list endpoints with `page` and `limit` query params.
- Return correct status codes: `200`, `201`, `204`, `400`, `401`, `403`, `404`, `409`.

## Security

- No hardcoded secrets — use `@nestjs/config` + env vars.
- All input validated via DTOs with `whitelist: true` and `forbidNonWhitelisted: true`.
- Use parameterized queries only — never interpolate user input.

## Database

- Use the ORM already in the project. No raw SQL without justification.
- Schema changes via migrations only.
- Use transactions for multi-table writes.
- Add indexes for columns in WHERE, JOIN, ORDER BY.

## Git

- Commit messages: imperative, concise — `Add order cancellation endpoint`.
- One logical change per commit.
- Branches: `feature/<desc>`, `fix/<desc>`, `chore/<desc>`.
- PRs must include tests and pass CI.

## Do NOT

- Add `console.log` or commented-out code.
- Add dependencies without approval.
- Skip tests.
- Use `any`.
- Write raw SQL without justification.
- Expose internal errors to clients.
