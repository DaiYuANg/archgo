package dbx

import "testing"

func BenchmarkNewStructMapperCached(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := NewStructMapper[auditedUser](); err != nil {
			b.Fatalf("NewStructMapper returned error: %v", err)
		}
	}
}

func BenchmarkStructMapperScanPlanCached(b *testing.B) {
	mapper := MustStructMapper[accountRecord]()
	columns := []string{"id", "nickname", "bio", "label"}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := mapper.scanPlan(columns); err != nil {
			b.Fatalf("scanPlan returned error: %v", err)
		}
	}
}

func BenchmarkStructMapperScanPlanAliasFallback(b *testing.B) {
	mapper := MustStructMapper[auditedUser]()
	columns := []string{`"users"."id"`, `"CREATED_BY"`, `"UPDATED_BY"`}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := mapper.scanPlan(columns); err != nil {
			b.Fatalf("scanPlan returned error: %v", err)
		}
	}
}

func BenchmarkMapperInsertAssignments(b *testing.B) {
	accounts := MustSchema("accounts", accountSchema{})
	mapper := MustMapper[accountRecord](accounts)
	entity := &accountRecord{
		Label: "ADMIN",
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := mapper.InsertAssignments(accounts, entity); err != nil {
			b.Fatalf("InsertAssignments returned error: %v", err)
		}
	}
}
