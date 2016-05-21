SET ThousandSep=',';
SET DecimalSep='.';
SET MoneyThousandSep=',';
SET MoneyDecimalSep='.';
SET MoneyFormat='$#,##0.00;($#,##0.00)';
SET TimeFormat='h:mm:ss TT';
SET DateFormat='DD.MM.YYYY';
SET TimestampFormat='DD.MM.YYYY h:mm:ss[.fff] TT';
SET FirstWeekDay=6;
SET BrokenWeeks=1;
SET ReferenceDay=0;
SET FirstMonthOfYear=1;
SET CollationLocale='en-US';
SET MonthNames='Jan;Feb;Mar;Apr;May;Jun;Jul;Aug;Sep;Oct;Nov;Dec';
SET LongMonthNames='January;February;March;April;May;June;July;August;September;October;November;December';
SET DayNames='Mon;Tue;Wed;Thu;Fri;Sat;Sun';
SET LongDayNames='Monday;Tuesday;Wednesday;Thursday;Friday;Saturday;Sunday';

LOAD
    @1 as Дата,
    Year(@1) as Год,
    Month(@1) as Месяц,
    Dual(Year(@1) & '-' & Month(@1), MonthStart(@1)) as 'Год и месяц',
    if(Len(@2) > 0, Lower(@2), 'Общие') as 'Член семьи',
    if(Wildmatch(@4, 'автобус', 'трамвай', 'троллейбус', 'маршрутка', 'метро', 'электричка', 'такси') > 0, 'транспорт', Lower(@3)) as Категория,
    Lower(@4) as Покупка,
    @5 as Цена
FROM [lib://qlikid_dddpaul1980/purchases-20160521.csv]
(txt, utf8, no labels, delimiter is ',', msq);
