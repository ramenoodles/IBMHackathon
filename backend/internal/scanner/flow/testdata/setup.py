class Tests(unittest.TestCase):
    def setUp(self):
        self.engine = create_engine(TEST_DATABASE_URL)
        self.session = Session(bind=self.engine)

    def test_foo(self):
        pass
