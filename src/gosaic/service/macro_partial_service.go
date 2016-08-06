package service

import (
	"fmt"
	"gosaic/model"
	"gosaic/util"
	"sync"

	"gopkg.in/gorp.v1"
)

type MacroPartialService interface {
	Service
	Insert(*model.MacroPartial) error
	Update(*model.MacroPartial) error
	Delete(*model.MacroPartial) error
	Get(int64) (*model.MacroPartial, error)
	GetOneBy(string, interface{}) (*model.MacroPartial, error)
	ExistsBy(string, interface{}) (bool, error)
	Count() (int64, error)
	CountBy(string, interface{}) (int64, error)
	Find(*model.Macro, *model.CoverPartial) (*model.MacroPartial, error)
	Create(*model.Macro, *model.CoverPartial) (*model.MacroPartial, error)
	FindOrCreate(*model.Macro, *model.CoverPartial) (*model.MacroPartial, error)
	FindMissing(*model.Macro, string, int, int) ([]*model.CoverPartial, error)
}

type macroPartialServiceImpl struct {
	dbMap *gorp.DbMap
	m     sync.Mutex
}

func NewMacroPartialService(dbMap *gorp.DbMap) MacroPartialService {
	return &macroPartialServiceImpl{dbMap: dbMap}
}

func (s *macroPartialServiceImpl) DbMap() *gorp.DbMap {
	return s.dbMap
}

func (s *macroPartialServiceImpl) Register() error {
	s.DbMap().AddTableWithName(model.MacroPartial{}, "macro_partials").SetKeys(true, "id")
	return nil
}

func (s *macroPartialServiceImpl) Insert(macroPartial *model.MacroPartial) error {
	err := macroPartial.EncodePixels()
	if err != nil {
		return err
	}
	return s.DbMap().Insert(macroPartial)
}

func (s *macroPartialServiceImpl) Update(macroPartial *model.MacroPartial) error {
	err := macroPartial.EncodePixels()
	if err != nil {
		return err
	}
	_, err = s.DbMap().Update(macroPartial)
	return err
}

func (s *macroPartialServiceImpl) Delete(macroPartial *model.MacroPartial) error {
	_, err := s.DbMap().Delete(macroPartial)
	return err
}

func (s *macroPartialServiceImpl) Get(id int64) (*model.MacroPartial, error) {
	macroPartial, err := s.DbMap().Get(model.MacroPartial{}, id)
	if err != nil {
		return nil, err
	} else if macroPartial == nil {
		return nil, nil
	}
	mp, ok := macroPartial.(*model.MacroPartial)
	if !ok {
		return nil, fmt.Errorf("Received struct is not a MacroPartial")
	}
	err = mp.DecodeData()
	if err != nil {
		return nil, err
	}
	return mp, nil
}

func (s *macroPartialServiceImpl) GetOneBy(column string, value interface{}) (*model.MacroPartial, error) {
	var macroPartial model.MacroPartial
	err := s.DbMap().SelectOne(&macroPartial, "select * from macro_partials where "+column+" = ? limit 1", value)
	if err != nil {
		return nil, err
	}

	err = macroPartial.DecodeData()
	if err != nil {
		return nil, err
	}

	return &macroPartial, err
}

func (s *macroPartialServiceImpl) ExistsBy(column string, value interface{}) (bool, error) {
	count, err := s.DbMap().SelectInt("select 1 from macro_partials where "+column+" = ? limit 1", value)
	return count == 1, err
}

func (s *macroPartialServiceImpl) Count() (int64, error) {
	return s.DbMap().SelectInt("select count(*) from macro_partials")
}

func (s *macroPartialServiceImpl) CountBy(column string, value interface{}) (int64, error) {
	return s.DbMap().SelectInt("select count(*) from macro_partials where "+column+" = ?", value)
}

func (s *macroPartialServiceImpl) doFind(macro *model.Macro, coverPartial *model.CoverPartial) (*model.MacroPartial, error) {
	p := model.MacroPartial{
		MacroId:        macro.Id,
		CoverPartialId: coverPartial.Id,
	}

	err := s.DbMap().SelectOne(&p, "select * from macro_partials where macro_id = ? and cover_partial_id = ?", p.MacroId, p.CoverPartialId)
	if err != nil {
		return nil, err
	}

	err = p.DecodeData()
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *macroPartialServiceImpl) Find(macro *model.Macro, coverPartial *model.CoverPartial) (*model.MacroPartial, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return s.doFind(macro, coverPartial)
}

func (s *macroPartialServiceImpl) doCreate(macro *model.Macro, coverPartial *model.CoverPartial) (*model.MacroPartial, error) {
	p := model.MacroPartial{
		MacroId:        macro.Id,
		CoverPartialId: coverPartial.Id,
		AspectId:       coverPartial.AspectId,
	}

	pixels, err := util.GetPartialLab(macro, coverPartial)
	if err != nil {
		return nil, err
	}
	p.Pixels = pixels

	err = p.EncodePixels()
	if err != nil {
		return nil, err
	}

	err = s.Insert(&p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *macroPartialServiceImpl) Create(macro *model.Macro, coverPartial *model.CoverPartial) (*model.MacroPartial, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return s.doCreate(macro, coverPartial)
}

func (s *macroPartialServiceImpl) FindOrCreate(macro *model.Macro, coverPartial *model.CoverPartial) (*model.MacroPartial, error) {
	s.m.Lock()
	defer s.m.Unlock()

	p, err := s.doFind(macro, coverPartial)
	if err == nil {
		return p, nil
	}

	// or create
	return s.doCreate(macro, coverPartial)
}

func (s *macroPartialServiceImpl) FindMissing(macro *model.Macro, order string, limit, offset int) ([]*model.CoverPartial, error) {
	s.m.Lock()
	defer s.m.Unlock()

	sql := `
select * from cover_partials
where not exists (
	select 1 from macro_partials
	where macro_partials.macro_id = ?
	and macro_partials.cover_partial_id = cover_partials.id
)
order by ?
limit ?
offset ?
`

	var coverPartials []*model.CoverPartial
	_, err := s.dbMap.Select(&coverPartials, sql, macro.Id, order, limit, offset)

	return coverPartials, err
}
