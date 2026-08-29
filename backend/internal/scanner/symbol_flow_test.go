package scanner

import "testing"

func TestExtractSymbolStepsPythonSetUp(t *testing.T) {
	src := `class Tests(unittest.TestCase):
    def setUp(self):
        self.engine = create_engine(TEST_DATABASE_URL)
        self.session = Session(bind=self.engine)

    def test_foo(self):
        pass
`
	steps := ExtractSymbolSteps(src, "test.py", "setUp", "python")
	if len(steps) < 3 {
		t.Fatalf("expected at least 3 steps, got %d", len(steps))
	}
	if steps[0].Kind != "entry" || steps[0].Label != "setUp()" {
		t.Fatalf("unexpected entry: %+v", steps[0])
	}
	if steps[1].Label != "create_engine()" {
		t.Fatalf("expected create_engine(), got %s", steps[1].Label)
	}
	if steps[2].Label != "Session()" {
		t.Fatalf("expected Session(), got %s", steps[2].Label)
	}
}
